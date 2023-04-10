package xbot

import (
	"time"
)

type BotOpts struct {
	spawn   bool
	panicBy BotPanicType

	// GetElementAttr
	ElemIndex int
	Attr      string
	Property  string

	//
	CaseInsensitive bool

	// Scroll options
	OffsetToTop float64

	// Input
	Submit bool

	// Timeout by seconds
	Timeout time.Duration

	// proxyLine with format `host:port:username:password:<OTHER>`
	proxyLine string

	BotCfg *BotConfig

	sleepSecBeforeAction float64
	scrollAsHuman        bool

	retry int
}

type BotOptFunc func(o *BotOpts)

func BotSpawn(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.spawn = b
	}
}

func WithBotConfig(cfg *BotConfig) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg = cfg
	}
}

// BotHeadless is not used, we use file `.rod:show` to control Headless or not
func BotHeadless(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.Headless = b
	}
}

func BotHighlight(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.Highlight = b
	}
}

func BotUserAgent(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.UserAgent = s
	}
}

func BotProxyLine(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.proxyLine = s
	}
}

func BotScreen(i int) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.Screen = i
	}
}

func BotSteps(i int) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.Steps = i
	}
}

func ElemIndex(i int) BotOptFunc {
	return func(o *BotOpts) {
		o.ElemIndex = i
	}
}

func ElemAttr(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.Attr = s
	}
}

func ElemProperty(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.Property = s
	}
}

func ElemOffsetToTop(f float64) BotOptFunc {
	return func(o *BotOpts) {
		o.OffsetToTop = f
	}
}

func WithCaseInsensitive(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.CaseInsensitive = b
	}
}

func InputSubmit(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.Submit = b
	}
}

func BotTimeout(i time.Duration) BotOptFunc {
	return func(o *BotOpts) {
		o.Timeout = i
	}
}

func WithPanicBy(i BotPanicType) BotOptFunc {
	return func(o *BotOpts) {
		o.panicBy = i
	}
}

func BindBotOpts(opt *BotOpts, opts ...BotOptFunc) {
	for _, f := range opts {
		f(opt)
	}
}

func WithSleepSec(t float64) BotOptFunc {
	return func(o *BotOpts) {
		o.sleepSecBeforeAction = t
	}
}

func WithScrollAsHuman(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.scrollAsHuman = b
	}
}

func WithRetry(i int) BotOptFunc {
	return func(o *BotOpts) {
		o.retry = i
	}
}

func WithRemoteService(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.remoteServiceUrl = s
	}
}
