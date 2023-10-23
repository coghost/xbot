package xbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/coghost/xpretty"
	"github.com/coghost/xutil"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"github.com/thoas/go-funk"
)

func (b *Bot) PanicIfErr(err error) {
	switch b.panicBy {
	case PanicByDump:
		dump.P(err)
	case PanicByLogError:
		log.Error().Err(err).Msg("error of bot")
	case PanicByLogFatal:
		log.Fatal().Err(err).Msg("error of bot")
	case PanicByLogPanic:
		log.Panic().Err(err).Msg("error of bot")
	default:
		xutil.PanicIfErr(err)
	}
}

func (b *Bot) DisableImages(brw *rod.Browser) {
	router := brw.HijackRequests()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		if ctx.Request.Type() == proto.NetworkResourceTypeImage {
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()
}

// DisableResources will hijack all resources
func (b *Bot) DisableResources(brw *rod.Browser, resources ...string) {
	if len(resources) == 0 {
		return
	}
	router := brw.HijackRequests()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		for _, res := range resources {
			if funk.Contains(ctx.Request.URL().String(), res) {
				ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
				return
			}
		}
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()
}

func (b *Bot) HandleXHR(brw *rod.Browser, res string, cb func(a, b string)) {
	router := brw.HijackRequests()
	router.MustAdd(res, func(ctx *rod.Hijack) {
		if ctx.Request.Type() == proto.NetworkResourceTypeXHR {
			ctx.MustLoadResponse()
			body := ctx.Response.Body()
			uri := ctx.Request.URL().String()
			cb(uri, body)
		}
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()
}

func (b *Bot) SetTimeout() {
	b.longToSec = LongToSec * time.Second
	b.mediumToSec = MediumToSec * time.Second
	b.shortToSec = ShortToSec * time.Second
	b.NapToSec = NapToSec * time.Second
}

func (b *Bot) GetPage(url string) {
	b.Pg.Timeout(b.longToSec).MustNavigate(url)
	b.Pg.Timeout(b.longToSec).MustWaitLoad()
}

func (b *Bot) GetPageE(url string) error {
	if e := b.Pg.Timeout(b.longToSec).Navigate(url); e != nil {
		return e
	}

	return b.Pg.Timeout(b.longToSec).WaitLoad()
}

func (b *Bot) CurrentUrl() string {
	return b.Pg.MustInfo().URL
}

// RetryWhenPanic
//
// args: tries, delay, showLogOrNot
// check `xutil.EnsureByRetry` for more detail
func (b *Bot) RetryWhenPanic(fn func(), args ...int) (int, error) {
	return xutil.EnsureByRetry(
		func() error {
			return rod.Try(func() {
				fn()
			})
		},
		args...,
	)
}

// CatchPanic just a wrapper of rod.Try.
//
// WARN: this only catch panic, takes no effect of error
//
// useful for bot.MustXxx
func (b *Bot) CatchPanic(fn func()) error {
	return rod.Try(fn)
}

// CatchPanicWithFb if fn panic, then fallback to fb.
//
// WARN: this only catch panic, takes no effect of error
//
// useful for bot.MustXxx
func (b *Bot) CatchPanicWithFb(fn func(), fb func() error) (err error) {
	err = rod.Try(fn)
	if errors.Is(err, context.DeadlineExceeded) {
		return fb()
	}
	return
}

func (b *Bot) BindPopovers(p []string) {
	b.popovers = p
}

// CloseIfHasPopovers
//
// will close all popovers bind to bot
// - try close by click the element
// - if failed, will try by Press Escape
//
// return total closed popovers
func (b *Bot) CloseIfHasPopovers() (hit int) {
	if len(b.popovers) == 0 {
		return
	}
	for _, sel := range b.popovers {
		hit += b.ClosePopover(sel)
	}
	if hit != 0 {
		log.Debug().Int("count", hit).Msg("not interactive, closed popovers")
	}
	return
}

func (b *Bot) ClosePopover(sel string) (hit int) {
	elems, err := b.Pg.Elements(sel)
	if err != nil {
		log.Error().Err(err).Str("selector", sel).Msg("find")
		return
	}
	if len(elems) == 0 {
		log.Trace().Msg("no popovers found")
		return
	}

	for _, elem := range elems {
		log.Debug().Str("popover", sel).Msg("try close")
		if !elem.MustInteractable() {
			elem.Overlay("popover is not interactable")
			return
		}

		b.Highlight(elem)
		e := elem.Click(proto.InputMouseButtonLeft, clickButtonTimes)
		if e != nil {
			log.Error().Err(e).Msg("close by left click")
		}
		hit += 1
	}
	return
}

// ClickPopoverByEsc close a popover by pressing escape
// In many cases, can use CloseIfHasPopovers instead
func (b *Bot) ClickPopoverByEsc(selector string, opts ...BotOptFunc) {
	if selector == "" {
		return
	}

	opt := BotOpts{ElemIndex: 0}
	BindBotOpts(&opt, opts...)
	elem := b.GetElem(selector, ElemIndex(opt.ElemIndex))
	if elem != nil {
		log.Debug().Str("popover", selector).Msg("Found Popover")
		b.Highlight(elem)
		elem.Timeout(b.NapToSec).MustKeyActions().Press(input.Escape).MustDo()
	}
}

func (b *Bot) MustPressEscape(sel string, opts ...BotOptFunc) {
	err := b.PressEscape(sel, opts...)
	b.PanicIfErr(err)
}

func (b *Bot) PressEscape(sel string, opts ...BotOptFunc) (err error) {
	if elem := b.GetElem(sel, opts...); elem != nil {
		// elem.MustKeyActions().Press(input.Escape).MustDo()
		elem.Timeout(b.shortToSec).MustKeyActions().Press(input.Escape).MustDo()
		return nil
	}
	return
}

func (b *Bot) PressTab(sel string, opts ...BotOptFunc) (err error) {
	if elem := b.GetElem(sel, opts...); elem != nil {
		xutil.RandSleep(0.5, 0.51)
		elem.MustKeyActions().Press(input.Tab).MustDo()
		return nil
	}
	return
}

func (b *Bot) ensureHighlight(elem *rod.Element) {
	b.ScrollToElem(elem, BotSteps(1), WithBotConfig(defaultCfg))

	if !elem.MustInteractable() {
		b.CloseIfHasPopovers()
	}

	b.Highlight(elem)
}

// EnsureAnyElem return the match elem
func (b *Bot) EnsureAnyElem(selectors ...string) (sel string, err error) {
	err = rod.Try(func() {
		r := b.Pg.Timeout(b.mediumToSec).Race()
		for _, s := range selectors {
			b.appendToRace(s, &sel, r)
		}
		r.MustDo()
	})
	return
}

// appendToRace:
// if directly add race.Element in EnsureAnyElem, will always return the
// last of the selectors
func (b *Bot) appendToRace(s string, sel *string, race *rod.RaceContext) {
	if funk.Contains(s, SEP) {
		ss := strings.Split(s, SEP)
		txt := strings.Join(ss[1:], SEP)
		race.ElementR(ss[0], txt).MustHandle(func(_ *rod.Element) {
			*sel = s
		})
	} else {
		race.Element(s).MustHandle(func(_ *rod.Element) {
			*sel = s
		})
	}
}

func (b *Bot) MustEnsureAnyElem(selectors ...string) string {
	start := time.Now()

	s, err := b.EnsureAnyElem(selectors...)
	b.PanicIfErr(err)

	cost := mathutil.ElapsedTime(start)
	log.Trace().Str("selector", s).Str("cost", cost).Msg("Ensure")
	return s
}

func (b *Bot) MustEnsureUrlHas(s string, opts ...BotOptFunc) {
	e := b.EnsureUrlHas(s, opts...)
	b.PanicIfErr(e)
}

func (b *Bot) EnsureUrlHas(s string, opts ...BotOptFunc) (err error) {
	opt := BotOpts{Timeout: MediumToSec}
	BindBotOpts(&opt, opts...)

	script := fmt.Sprintf(`() => decodeURIComponent(window.location.href).includes("%s")`, s)
	err = rod.Try(func() {
		b.Pg.Timeout(time.Second * time.Duration(opt.Timeout)).MustWait(script).CancelTimeout()
	})

	if err != nil {
		log.Error().Err(err).Msg(xpretty.Yellowf("Fail: %s", script))
	}

	return err
}

// MustEval
//
// a wrapper with MediumTo to rod.Page.MustEval
//
// if you want to get error, please use rod.Page.Eval instead
func (b *Bot) MustEval(script string) (res string) {
	res = b.Pg.Timeout(b.mediumToSec).MustEval(script).String()
	return res
}

func (b *Bot) MustFillBar(sel, text string, opts ...BotOptFunc) (txt string) {
	txt, err := b.FillBar(sel, text, opts...)
	b.PanicIfErr(err)
	return txt
}

func (b *Bot) FillBar(sel, text string, opts ...BotOptFunc) (txt string, err error) {
	opt := BotOpts{Submit: false}
	BindBotOpts(&opt, opts...)

	// elem := b.Pg.Timeout(time.Second * b.ShortTo).MustElement(sel).CancelTimeout()
	elem := b.GetElem(sel, opts...)
	if elem == nil {
		return "", ErrorSelNotFound
	}

	b.CloseIfHasPopovers()
	b.Highlight(elem)
	// elem = elem.Timeout(time.Second * b.ShortTo).MustSelectAllText().MustInput(text)
	elem = b.FillAsHuman(elem, text)
	if opt.Submit {
		xutil.RandSleep(0.1, 0.15)
		// elem = elem.MustPress(input.Enter)
		elem.MustKeyActions().Press(input.Enter).MustDo()
		// return nil
	}
	// just try to get text, won't matter if fails
	txt, _ = elem.Text()
	elem.CancelTimeout()
	return
}

// FillAsHuman
//
//	each time before enter (n=args[0] or 5) chars, we wait (to=args[1]/10 or 0.1) seconds
//
//	@return *rod.Element
func (b *Bot) FillAsHuman(elem *rod.Element, text string, args ...int) *rod.Element {
	elem.MustSelectAllText().MustInput("")
	n := xutil.FirstOrDefaultArgs(0, args...)
	if n == 0 {
		n = xutil.AorB(b.Config.PerInputLength, 5)
	}

	arr := xutil.NewStringSlice(text, n, true)
	for _, str := range arr {
		e := elem.Input(str)
		b.PanicIfErr(e)
	}

	to := 0.1
	if len(args) >= 2 {
		to = cast.ToFloat64(args[1]) / 10
	}
	xutil.RandSleep(to-0.01, to+0.01)
	return elem
}

func (b *Bot) FillCharsOneByOne(elem *rod.Element, value string) {
	elem.MustKeyActions().Type([]input.Key(fmt.Sprintf("%v", value))...).MustDo()
}

// MGetElems
//
// get all elems if found by selectors
func (b *Bot) MGetElems(selectors []string, opts ...BotOptFunc) (elems []*rod.Element) {
	for _, sel := range selectors {
		b.GetElem(sel, opts...)
		e1 := b.GetElems(sel)
		elems = append(elems, e1...)
	}
	return
}

// MGetElemsAllAttr
//
// get all elems' attribute
func (b *Bot) MGetElemsAllAttr(selectors []string, opts ...BotOptFunc) []string {
	var attrs []string
	for _, elem := range b.MGetElems(selectors, opts...) {
		at := b.getElemAttr(elem, opts...)
		attrs = append(attrs, at)
	}
	return attrs
}

// GetElems
//
// as the document of go-rod:
// If a multi-selector doesn't find anything, it will immediately return an empty list.
//
// and as the test results of `func (s *botSuite) TestGetElems()`
// the whole GetElems' time cost is less than 0.2 second
//
// get all elements that match the css selector or []
// just an alias of bot.Page.Elements
//
// if you want handle the error info, please call b.Pg.Elements directly
func (b *Bot) GetElems(selector string, opts ...BotOptFunc) (elems []*rod.Element) {
	if selector == "" {
		return
	}

	if strings.Contains(selector, SEP) {
		log.Error().Str("selector", selector).Msgf("invalid format which contains %q", SEP)
		return
	}

	opt := BotOpts{Timeout: 0}
	BindBotOpts(&opt, opts...)
	if opt.Timeout != 0 {
		b.GetElem(selector, BotTimeout(opt.Timeout))
	}

	elems, err := b.Pg.Elements(selector)
	if err != nil {
		log.Error().Err(err).Str("selector", selector).Msg("error of GetElems")
	}

	return elems
}

func (b *Bot) GetElemWithoutDelay(selector string, indexArgs ...int) *rod.Element {
	index := xutil.FirstOrDefaultArgs(0, indexArgs...)
	elems := b.GetElems(selector)

	if funk.IsEmpty(elems) {
		return nil
	}

	if index < 0 {
		index = len(elems) + index
	}
	index = xutil.Min(index, len(elems)-1)
	return elems[index]
}

// GetElemUntilInteractable
//
// in most cases, this is not needed
// for now, tested with popovers, when some site show popovers at a random time window
func (b *Bot) GetElemUntilInteractable(selector string, opts ...BotOptFunc) (elem *rod.Element) {
	ts := time.Now()
	for {
		elem = b.GetElem(selector, opts...)
		if elem == nil {
			xutil.RandSleep(0.5, 1)
			continue
		}

		if elem.MustInteractable() {
			return
		}

		log.Warn().Bool("interactable", elem.MustInteractable()).Msgf("un-interactable of %q", selector)
		xutil.RandSleep(0.5, 1)

		cost := xutil.ElapsedSeconds(ts, 2)
		if cost > MediumToSec {
			return nil
		}
	}
}

// GetElem by default wait (MediumTo) for the element to appear and return it
//
// use cases:
//  1. opt.Timeout == 0
//     same as GetElemWithoutDelay
//  2. opt.Timeout != 0
//     2.1 no index passed
//     - wait the element for opt.Timeout
//     - and return the found elem or nil
//     2.2 with index (support python style index: -1, ...)
//     - wait the element for opt.Timeout
//     - re-get elem by GetElemWithoutDelay
//     2.3 WARN: if SEP(@@@) in selector
//     - will return the elem found by ElementR
//     - will skip the index passed in
func (b *Bot) GetElem(selector string, opts ...BotOptFunc) (elem *rod.Element) {
	if selector == "" {
		w, i := xpretty.Caller(2)
		log.Warn().Str("file", w).Int("line", i).Msg("selector is empty")
		return
	}
	b.selector = selector

	byText := strings.Contains(selector, SEP)

	opt := BotOpts{ElemIndex: xutil.MaxInt, Timeout: MediumToSec, root: b.root}
	BindBotOpts(&opt, opts...)

	if !byText && int(opt.Timeout) == 0 {
		return b.GetElemWithoutDelay(selector, opt.ElemIndex)
	}

	ts := time.Now()
	// xutil.DumbLog(fmt.Sprintf("wait for (%s) to appear", selector))
	// wait elem of selector to appear

	dur := time.Duration(opt.Timeout) * time.Second
	var err error
	if funk.Contains(selector, SEP) {
		ss := strutil.Split(selector, SEP)
		txt := ss[len(ss)-1]
		if len(ss) == 3 {
			// when selector is like div.abc@@@---@@@txt, we use exactly match
			m := "/^%s$/"
			if opt.CaseInsensitive {
				m += "i"
			}
			txt = fmt.Sprintf(m, txt)
		}
		if opt.root != nil {
			elem, err = opt.root.Timeout(dur).ElementR(ss[0], txt)
		} else {
			elem, err = b.Pg.Timeout(dur).ElementR(ss[0], txt)
		}
	} else {
		if opt.root != nil {
			elem, err = opt.root.Timeout(dur).Element(selector)
		} else {
			elem, err = b.Pg.Timeout(dur).Element(selector)
		}
	}

	if err != nil {
		return elem
	}

	nonMax := opt.ElemIndex != xutil.MaxInt
	nonZero := opt.ElemIndex != 0
	// if specify index not 0/max, and no SEP(@@@) in selector, re-get it by GetElems
	// this is used when we first need to wait elems to appear, then get the specified one
	if !byText && nonMax && nonZero {
		elem = b.GetElemWithoutDelay(selector, opt.ElemIndex)
	}

	if v := elem.MustInteractable(); !v {
		log.Trace().Msgf("[GetElem] %s interactable=%t", selector, v)
	}
	cost := xutil.ElapsedSeconds(ts, 2)
	if cost > 2.0 {
		log.Debug().Float64("cost", cost).Str("selector", selector).Msg("GetElem")
	}

	return elem
}

// GetElemWithRetry
//
// works with Panic, not error
func (b *Bot) GetElemWithRetry(selector string, retryTimes int, opts ...BotOptFunc) (elem *rod.Element, err error) {
	opt := BotOpts{Timeout: NapToSec}
	BindBotOpts(&opt, opts...)

	i, err := b.RetryWhenPanic(func() {
		elem = b.GetElem(selector, opts...)
		if elem == nil {
			panic("failed of get element")
		}
	}, retryTimes)
	if err != nil {
		log.Debug().Msgf("though tried %d times, failed with %v", i, err)
	}
	return
}

func (b *Bot) GetElementAttrByRetry(selector string, opts ...BotOptFunc) string {
	opt := BotOpts{
		retry: 3,
	}
	BindBotOpts(&opt, opts...)

	var v string
	for i := 0; i < opt.retry; i++ {
		v = b.GetElementAttr(selector, opts...)
		if strings.TrimSpace(v) != "" {
			return v
		}
		xutil.RandSleep(0.5, 0.6)
	}
	return v
}

// GetElementAttr by default return the innerText of given selector
//
// but can be customized by ElemAttr("attr_value")
//
// - will panic is selector is ""
// - will return "" if no elem found by given selector
func (b *Bot) GetElementAttr(selector string, opts ...BotOptFunc) string {
	if selector == "" {
		panic("selector is empty")
	}

	elem := b.GetElem(selector, opts...)
	if elem == nil {
		return ""
	}

	b.Highlight(elem)
	return b.getElemAttr(elem, opts...)
}

func (b *Bot) GetElemAttr(elem *rod.Element, opts ...BotOptFunc) string {
	return b.getElemAttr(elem, opts...)
}

func (b *Bot) getElemAttr(elem *rod.Element, opts ...BotOptFunc) (val string) {
	opt := BotOpts{Attr: "innerText"}
	BindBotOpts(&opt, opts...)

	attr := opt.Attr
	if attr == "" || attr == "innerText" {
		return elem.MustText()
	}

	s, e := elem.Attribute(attr)
	log.Trace().Str("Attr", attr).Msg("will try get")

	if e != nil {
		log.Error().Err(e).Str("Attr", attr).Msg("getElemAttr")
		return
	}

	if s == nil {
		log.Debug().Str("Attr", attr).Msg("get NIL of")
		return
	}

	return *s
}

func (b *Bot) GetElementProp(selector string, opts ...BotOptFunc) (string, error) {
	opt := BotOpts{
		ElemIndex: 0,
		Property:  "value",
	}
	BindBotOpts(&opt, opts...)

	elem := b.GetElem(selector, ElemIndex(opt.ElemIndex))
	b.Highlight(elem)
	s, e := elem.Property(opt.Property)
	if e == nil {
		s1 := s.String()
		return s1, e
	} else {
		return "", e
	}
}

func (b *Bot) GetWindowInnerHeight() float64 {
	h := b.Pg.Timeout(b.shortToSec).MustEval(`() => window.innerHeight`).Int()
	return float64(h)
}

func (b *Bot) GetScrollHeight() float64 {
	h := b.Pg.Timeout(b.shortToSec).MustEval(`() => document.body.scrollHeight`).Int()
	return float64(h)
}

func (b *Bot) MustScrollAndClick(selector interface{}, opts ...BotOptFunc) {
	err := b.ScrollAndClick(selector, opts...)
	b.PanicIfErr(err)
}

// ScrollAndClick
//
// if you pass elem here, please remember the Timeout you used to get the elem
// it will be passed through until cancel called
func (b *Bot) ScrollAndClick(selector interface{}, opts ...BotOptFunc) error {
	if funk.IsEmpty(selector) {
		return fmt.Errorf("empty selector(%v) found", selector)
	}
	elem := b.RecalculateElem(selector, opts...)
	if elem == nil {
		return ErrorSelNotFound
	}
	return b.ScrollAndClickElem(elem)
}

func (b *Bot) ClickAndSwitchToNewPage(selector interface{}, opts ...BotOptFunc) (*rod.Page, error) {
	opt := BotOpts{
		Timeout: MediumToSec,
	}
	BindBotOpts(&opt, opts...)

	wait := b.Pg.Timeout(time.Second * time.Duration(opt.Timeout)).WaitOpen()
	if err := b.ScrollAndClick(selector, opts...); err != nil {
		return nil, err
	}
	pg, err := wait()
	b.UpdatePage(pg)
	return pg, err
}

func (b *Bot) MustClickAndSwitchToNewPage(selector interface{}, opts ...BotOptFunc) *rod.Page {
	pg, err := b.ClickAndSwitchToNewPage(selector, opts...)
	if err != nil {
		b.PanicIfErr(err)
	}
	return pg
}

func (b *Bot) MustClickAndSwitchToNewPageWithScript(selector interface{}, opts ...BotOptFunc) *rod.Page {
	wait := b.Pg.WaitOpen()
	elem := b.GetElem(selector.(string), opts...)
	_ = b.ClickWithScript(elem)
	pg, err := wait()
	b.UpdatePage(pg)

	if err != nil {
		b.PanicIfErr(err)
	}
	return pg
}

func (b *Bot) MustScrollAndClickElem(elem *rod.Element, retryArgs ...uint) {
	err := b.ScrollAndClickElem(elem, retryArgs...)
	b.PanicIfErr(err)
}

// ScrollAndClickElem
//
// please remember go-rod's element's Timeout will be passed through until cancel called
func (b *Bot) ScrollAndClickElem(elem *rod.Element, retryArgs ...uint) error {
	var attempt uint = 4
	if len(retryArgs) > 0 {
		attempt = retryArgs[0]
	}

	tried := 1
	err := retry.Do(
		func() error {
			return rod.Try(func() {
				if tried > 1 {
					log.Debug().Uint("total", attempt).Int("tried", tried).Msg("scroll and click")
				}
				tried += 1
				if err := b.ScrollAndClickOnce(elem); err != nil {
					panic(err)
				}
			})
		},
		retry.Attempts(attempt),
	)
	return err
}

func (b *Bot) ScrollAndClickOnce(elem *rod.Element) (err error) {
	if elem == nil {
		xutil.DumpCallerStack()
		panic("elem is nil")
	}

	if v := elem.MustInteractable(); !v {
		n := b.CloseIfHasPopovers()
		if n > 0 {
			log.Debug().Bool("clickable", v).Msg("elem un-clickable try close popovers")
		}
	}

	return b.ClickElemAndFbWithJs(elem)
}

func (b *Bot) ClickElemAndFbWithJs(elem *rod.Element) (err error) {
	err = b.CatchPanicWithFb(
		func() {
			if e := b.ClickElem(elem); e != nil {
				panic(e)
			}
		}, func() error {
			return b.ClickWithScript(elem)
		})
	return err
}

func (b *Bot) MustClickElem(elem *rod.Element) {
	e := b.ClickElem(elem)
	b.PanicIfErr(e)
}

func (b *Bot) ClickElem(elem *rod.Element, highlight ...bool) error {
	if len(highlight) == 0 {
		b.ensureHighlight(elem)
	}
	e := elem.Timeout(b.shortToSec).Click(proto.InputMouseButtonLeft, clickButtonTimes)
	if e != nil {
		log.Warn().Interface("selector", b.selector).Err(e).Msg("Err: close by left click")
	}
	return e
}

func (b *Bot) MustClickWithScript(elem *rod.Element, args ...int) {
	e := b.ClickWithScript(elem, args...)
	b.PanicIfErr(e)
}

// ClickWithScript
//
// can skip Highlight be passing args with nonZero value
func (b *Bot) ClickWithScript(elem *rod.Element, args ...int) error {
	v := xutil.FirstOrDefaultArgs(0, args...)
	if v == 0 {
		b.ensureHighlight(elem)
	}
	to := b.shortToSec
	if len(args) >= 2 {
		to = time.Duration(args[1]) * time.Second
	}

	_, e := elem.Timeout(to).Eval(` (elem) => { this.click() }`, elem)
	if e != nil {
		log.Error().Err(e).Msg("Err: close by this.click()")
		return e
	}

	_, ei := elem.Interactable()
	if errors.Is(ei, &rod.ErrObjectNotFound{}) {
		return nil
	}
	return ei
}

// Highlight
//
// similar with elem.Overlay
func (b *Bot) Highlight(elem *rod.Element) {
	show, hide := 0.333, 0.2
	style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`

	go b._highlight(elem, show, hide, style, 0)
}

// HighlightBlink
//
//   - style: default is "box-shadow: 0px 0px 5px 5px rgba(128, 0, 0, 1);"
//   - count: blink how many times
func (b *Bot) HighlightBlink(elem *rod.Element, count int, style string) {
	show, hide := 0.25, 0.15
	// style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.15s ease-in-out; animation-delay: 0.1s;`
	// style = `box-shadow: 0px 0px 10px 8px #c8c8c8, 0px 0px 12px 10px #a0a0a0; transition: all 0.15s ease-in-out; animation-delay: 0.1s;`
	const styleLayout = "%s transition: all 0.15s ease-in-out; animation-delay: 0.1s;"
	style_ := `box-shadow: 0px 0px 5px 5px rgba(128, 0, 0, 1);`
	if style == "" {
		style = style_
	}
	b._highlight(elem, show, hide, fmt.Sprintf(styleLayout, style), count)
}

func (b *Bot) _highlight(elem *rod.Element, show, hide float64, style string, count int) (cost float64) {
	ts := time.Now()
	if !b.Config.Highlight {
		return
	}
	if elem == nil {
		return
	}

	ob, err := elem.Eval(`e => {return this.getAttribute("style")}`)
	if err != nil {
		log.Debug().Msg("No style found")
		return
	}
	origStyle := ob.Value.String()
	// origStyle := elem.MustEval(`e => {return this.getAttribute("style")}`).String()
	// style := `box-shadow: rgb(255, 156, 85) 0px 0px 0px 8px, rgb(255, 85, 85) 0px 0px 0px 10px; transition: all 0.5s ease-in-out; animation-delay: 0.1s;`
	// show, hide := 0.333, 0.2
	v := 0.05
	if count == 0 {
		count = xutil.AorB(b.Config.HighlightTimes, 3)
	}

	for i := 0; i < count; i++ {
		script := fmt.Sprintf(`() => this.setAttribute("style", "%s");`, style)
		_, _ = elem.Eval(script)
		xutil.RandSleep(show-v, show+v)
		script = fmt.Sprintf(`() => this.setAttribute("style", "%s");`, origStyle)
		_, _ = elem.Eval(script)
		xutil.RandSleep(hide-v, hide+v)
	}

	cost = xutil.ElapsedSeconds(ts, 2)
	return
}

// ScrollToElem
//
// though rod's built with support of scroll to elem before click/input
// we want to control the scroll manually, to behavior more like human
func (b *Bot) ScrollToElem(elem *rod.Element, opts ...BotOptFunc) {
	if b.Config == nil {
		b.Config = defaultCfg
	}
	oft := xutil.IfaceAorB(b.Config.OffsetToTop, 0.25).(float64)
	opt := BotOpts{OffsetToTop: oft}
	BindBotOpts(&opt, opts...)

	h := b.GetWindowInnerHeight()
	box, err := b.GetElemBox(elem)
	if err != nil {
		log.Debug().Err(err).Msg("cannot get elem box")
	}
	scrollDistance := box.Top - h*opt.OffsetToTop
	dist := xutil.IfaceAorB(b.Config.ScrollDistanceBase, 640.0).(float64)
	// take how many steps to scroll elem to position
	steps := xutil.AorB(b.Config.Steps, 32)
	steps = int((scrollDistance / dist) * float64(steps))

	log.Trace().Float64("distance", scrollDistance).Int("steps", steps).Msg("Will scroll with")
	e := b.Pg.Mouse.Scroll(0.0, scrollDistance, steps)
	b.PanicIfErr(e)
}

func (b *Bot) ScrollToElemDirectly(elem *rod.Element) error {
	box, err := b.GetElemBox(elem)
	if err != nil {
		return err
	}

	return b.Pg.Mouse.Scroll(0.0, box.Top, 1)
}

func (b *Bot) MustScrollToXY(x, y float64) {
	b.Pg.Mouse.MustScroll(x, y)
}

func (b *Bot) MustScrollToTop() {
	h := b.GetScrollHeight()
	e := b.Pg.Mouse.Scroll(0, -h, 16 /* steps */)
	b.PanicIfErr(e)
}

func (b *Bot) MustScrollToBottom(opts ...BotOptFunc) {
	e := b.ScrollToBottom(opts...)
	b.PanicIfErr(e)
}

func (b *Bot) ScrollToBottom(opts ...BotOptFunc) error {
	opt := BotOpts{scrollAsHuman: true, BotCfg: NewDefaultBotCfg()}
	BindBotOpts(&opt, opts...)
	if v := opt.sleepSecBeforeAction; v != 0 {
		xutil.RandSleep(v, v+0.5)
	}

	h := b.GetScrollHeight()
	// e := b.Pg.Mouse.Scroll(0, h, opt.Steps)
	e := b.ScrollLikeHuman(0, h, opts...)
	return e
}

func (b *Bot) GetElemBox(elem interface{}) (box Box, err error) {
	elem = b.RecalculateElem(elem)
	rect := "() => JSON.stringify(this.getBoundingClientRect())"
	err = rod.Try(func() {
		dat := elem.(*rod.Element).Timeout(b.shortToSec).MustEval(rect).String()
		log.Trace().Msg(dat)
		e := json.Unmarshal([]byte(dat), &box)
		if e != nil {
			log.Error().Err(e).Msg("get elem box failed")
		}
		b.PanicIfErr(e)
	})
	return
}

// RecalculateElem automatically decide GetElem/GetElement by type of elem
func (b *Bot) RecalculateElem(elem interface{}, opts ...BotOptFunc) (newElem *rod.Element) {
	switch elem := elem.(type) {
	case string:
		newElem = b.GetElem(elem, opts...)
	case *rod.Element:
		newElem = elem
	}
	return
}

// ScrollLikeHuman performs a scroll action like human,
//
//	scroll down a bit, then sleep a random mills
func (b *Bot) ScrollLikeHuman(offsetX, offsetY float64, opts ...BotOptFunc) error {
	page := b.Pg
	opt := BotOpts{scrollAsHuman: true, BotCfg: NewDefaultBotCfg()}
	BindBotOpts(&opt, opts...)

	steps := opt.BotCfg.Steps

	b.ScrollAsHuman = &ScrollAsHuman{
		enabled:          opt.scrollAsHuman,
		longSleepChance:  0.1,
		shortSleepChance: 0.2,
		scrollUpChance:   0.9,
	}

	if !b.ScrollAsHuman.enabled || steps == 0 {
		err := page.Mouse.Scroll(offsetX, offsetY, 1)
		xutil.RandSleep(0.1, 0.2)
		return err
	}

	tooSlowTimeoutSec := 20
	totalScrolled := 0.0
	totalNeeded := offsetY

	base := offsetY / float64(steps)

	if offsetY < 0 {
		totalNeeded = math.Abs(totalNeeded)
	}
	startAt := time.Now()

	for totalScrolled < totalNeeded {
		yNegative := false
		// handle too slow scroll
		cost := xutil.ElapsedSeconds(startAt, 2)
		if cost > float64(tooSlowTimeoutSec) {
			err := page.Mouse.Scroll(offsetX, totalNeeded-totalScrolled, 1)
			xutil.RandSleep(0.1, 0.2)
			return err
		}

		chance := rand.Float64()

		if chance < b.ScrollAsHuman.longSleepChance {
			xutil.RandSleep(0.5, 0.6)
			continue
		}

		if chance < b.ScrollAsHuman.shortSleepChance {
			xutil.RandSleep(0.25, 0.3)
			continue
		}

		distance := rand.Intn(10) + int(base)
		if chance > b.ScrollAsHuman.scrollUpChance {
			yNegative = true
			distance = rand.Intn(20) + int(base)*2
		}
		if v := totalNeeded - totalScrolled; int(v) < distance {
			distance = int(v)
		}
		if yNegative {
			distance = -distance
		}

		if e := page.Mouse.Scroll(offsetX, float64(distance), steps); e != nil {
			return e
		}
		totalScrolled += float64(distance)
	}

	return nil
}

func (b *Bot) UpdatePageE(page *rod.Page) error {
	b.PrevPage, b.Pg = b.Pg, page
	_, err := b.Pg.Activate()
	return err
}

func (b *Bot) UpdatePage(page *rod.Page) {
	err := b.UpdatePageE(page)
	b.PanicIfErr(err)
}

func (b *Bot) ResetToOriginalPage() {
	if b.PrevPage != nil && b.Pg != nil {
		b.Pg.MustClose()
	}
	b.Pg = b.PrevPage
}

func (b *Bot) BindIframe(frame *rod.Page) {
	b.Iframe = frame
}

func (b *Bot) ResetRoot() {
	b.root = nil
}

func (b *Bot) BindRoot(root *rod.Element) {
	b.root = root
}

func (b *Bot) SetPageWithCookies(page *rod.Page, raw string) error {
	var cookies []proto.NetworkCookie

	err := json.Unmarshal([]byte(raw), &cookies)
	if err != nil {
		return err
	}

	var nodes []*proto.NetworkCookieParam
	for _, cookie := range cookies {
		nodes = append(nodes, &proto.NetworkCookieParam{
			Name:         cookie.Name,
			Value:        cookie.Value,
			Domain:       cookie.Domain,
			Path:         cookie.Path,
			Secure:       cookie.Secure,
			HTTPOnly:     cookie.HTTPOnly,
			SameSite:     cookie.SameSite,
			Expires:      cookie.Expires,
			Priority:     cookie.Priority,
			SameParty:    cookie.SameParty,
			SourceScheme: cookie.SourceScheme,
			SourcePort:   &cookie.SourcePort,
		})
	}

	return page.SetCookies(nodes)
}

func (b *Bot) Close() {
	if b != nil && b.Brw != nil {
		b.Brw.Close()
	}
}
