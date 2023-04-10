package xbot_test

import (
	"context"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/coghost/xbot"
	"github.com/coghost/xlog"
	"github.com/coghost/xutil"
	"github.com/go-rod/rod"
	"github.com/remeh/sizedwaitgroup"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"
	"github.com/thoas/go-funk"
)

var swg = sizedwaitgroup.New(4)

type botSuite struct {
	suite.Suite
}

func TestBotNoBrw(t *testing.T) {
	suite.Run(t, new(botSuite))
}

func (s *botSuite) SetupSuite() {
	xlog.InitLog(xlog.WithLevel(1))
}

func (s *botSuite) TearDownTest() {
	swg.Wait()
}

type baseArgs struct {
	url string
	sel string

	img         string
	searchTerm  string
	nonExistSel string

	location        string
	locationSuggest string

	category        string
	categoryIndex   int
	confirmCategory string

	popovers []string

	has []string
	no  []string

	submit string
	urlHas string
}

var baidu = &baseArgs{
	url:         "https://www.baidu.com/",
	img:         `img[id="s_lg_img_new"]`,
	popovers:    []string{},
	submit:      `input[id="su"]`,
	searchTerm:  "input#kw",
	nonExistSel: "div.thisshouldneverbefoundinthehtml",
	has:         []string{"input#kw"},
	no:          []string{"div.thisshouldneverbefoundinthehtml"},
}

var jandan = &baseArgs{
	url:         "https://jandan.net/",
	img:         "div.post>div>a>img",
	searchTerm:  "",
	nonExistSel: "",
	popovers:    []string{},
	has:         []string{},
	no:          []string{},
	submit:      "",
	urlHas:      "",
}

var blocket = &baseArgs{
	url:             "https://jobb.blocket.se/",
	sel:             "",
	img:             "",
	searchTerm:      "input#whatinput",
	nonExistSel:     "div.thisshouldneverbefoundinthehtml",
	location:        "",
	locationSuggest: "",
	category:        "div.cat div.checkbox>label>b@@@Data & IT",
	categoryIndex:   2,
	confirmCategory: `div[data-type="filter"]@@@Data & IT`,
	popovers:        []string{"button#accept-ufti"},
	has:             []string{},
	no:              []string{},
	submit:          "a#search-button",
	urlHas:          "lediga-jobb",
}

type botTest struct {
	name string
	args *baseArgs

	fn func()
	fb func() error

	res  string
	sels []string

	tries int
	delay int
	to    int
	index int

	// duration time.Duration

	withSubmit bool

	want     int
	wantInt  int
	wantBool bool
	wantF64  float64
	wantStr  string
	wantErr  error
}

var testImages = []botTest{
	{
		name: "baidu",
		args: baidu,
	},
	{
		name: "jandan",
		args: jandan,
	},
}

func runWorker(tt botTest, handle func(b *xbot.Bot, tt botTest), args ...bool) {
	openPage := true
	if len(args) > 0 {
		openPage = args[0]
	}

	swg.Add()
	go func(tt botTest) {
		defer swg.Done()
		b := xbot.NewBot(xbot.BotScreen(1), xbot.BotHeadless(false), xbot.BotHighlight(true), xbot.BotUserAgent(xbot.UA))
		defer b.Brw.Close()
		if funk.NotEmpty(tt.args) && openPage {
			b.GetPage(tt.args.url)
		}

		handle(b, tt)
	}(tt)
}

func getImageSize(b *xbot.Bot, uri, sel string) int {
	bin := b.Pg.Timeout(time.Second * 30).MustElement(sel).MustResource()
	return len(bin)
}

func (s *botSuite) Test_01_DisableImages_Enabled() {
	for _, tt := range testImages {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			v := getImageSize(b, tt.args.url, tt.args.img)
			s.NotZero(v, tt.name)
			s.Greater(v, 100, tt.name)
		})
	}

}

func (s *botSuite) Test_01_DisableImages_Disabled() {
	for _, tt := range testImages {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			b.DisableImages(b.Brw)
			b.GetPage(tt.args.url)
			v := getImageSize(b, tt.args.url, tt.args.img)
			s.Zero(v, tt.name)
		}, false)
	}
}

func (s *botSuite) Test_01_DisableResourcesBySubStr_Enabled() {
	for _, tt := range testImages {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			v := getImageSize(b, tt.args.url, tt.args.img)
			s.NotZero(v, tt.name)
			s.Greater(v, 100, tt.name)
		})
	}
}

func (s *botSuite) Test_01_DisableResourcesBySubStr_Disabled() {
	for _, tt := range testImages {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			b.DisableResources(b.Brw, ".jpg", ".png", ".jpeg")
			defer b.Brw.MustClose()
			b.GetPage(tt.args.url)

			v := getImageSize(b, tt.args.url, tt.args.img)
			s.Zero(v, tt.name)
		}, false)
	}
}

func (s *botSuite) Test_02_HandleXHR() {
	tests := []botTest{
		{
			name:     "existed jandan dot resource",
			args:     jandan,
			res:      "*dot*",
			wantBool: true,
		},
		{
			name:     "unexisted jandan",
			args:     jandan,
			res:      "*unexisted_resource*",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			b.HandleXHR(b.Brw, tt.res,
				func(_, raw string) {
					// if we can run here, means match tt.res, raw could never be nil
					if tt.wantBool {
						s.NotNil(raw, tt.name)
						s.Greater(len(raw), 16, tt.name)
						s.Less(len(raw), 1024, tt.name)
					} else {
						panic("this should never be called")
					}
				})
			b.GetPage(tt.args.url)
			image := getImageSize(b, tt.args.url, tt.args.img)
			s.NotZero(image, tt.name)
		}, false)
	}
}

// get page only once
func (s *botSuite) Test_03_CatchOrFb() {
	// b is required here for predefined fns
	b := xbot.NewBot(xbot.BotScreen(1), xbot.BotHeadless(false), xbot.BotUserAgent(xbot.UA))
	defer b.Brw.Close()
	b.GetPage(baidu.url)

	fn1 := func() {
		b.Pg.Timeout(time.Second * 1).MustElement(baidu.nonExistSel)
	}
	fn2 := func() {
		b.Pg.Timeout(time.Second * 1).MustElement(baidu.searchTerm)
	}
	fb1 := func() error {
		_, err := b.Pg.Timeout(time.Second * 1).Element(baidu.nonExistSel)
		return err
	}
	fb2 := func() error {
		_, err := b.Pg.Timeout(time.Second * 1).Element(baidu.searchTerm)
		return err
	}

	tests1 := []botTest{
		{
			name:    "failed",
			fn:      fn1,
			wantErr: &rod.ErrTry{Value: context.DeadlineExceeded},
		},
		{
			name:    "found",
			fn:      fn2,
			wantErr: nil,
		},
	}

	tests2 := []botTest{
		{
			name:    "fn with fb error",
			wantErr: context.DeadlineExceeded,
			fn:      fn1,
			fb:      fb1,
		},
		{
			name:    "fn with fb success",
			wantErr: nil,
			fn:      fn1,
			fb:      fb2,
		},
	}

	for _, tt := range tests1 {
		got := b.CatchPanic(tt.fn)
		s.ErrorIs(got, tt.wantErr)
	}

	for _, tt := range tests2 {
		got := b.CatchPanicWithFb(tt.fn, tt.fb)
		s.Equal(tt.wantErr, got, tt.name)
	}
}

func (s *botSuite) Test_04_RetryWhenPanic() {
	pfn := func() {
		panic("panic func")
	}
	tests := []botTest{
		{
			name:    "no error with default 3",
			fn:      func() {},
			tries:   -1,
			want:    1,
			wantErr: nil,
		},
		{
			name:    "panic with default 3 times",
			fn:      pfn,
			tries:   -1,
			want:    3,
			wantStr: "All attempts fail:\n#1: error value: \"panic func\"",
		},
		{
			name:    "panic with 0 times",
			fn:      pfn,
			tries:   0,
			want:    0,
			wantErr: retry.Error(retry.Error{}),
		},
		{
			name:    "panic with 2 times",
			fn:      pfn,
			tries:   2,
			want:    2,
			wantStr: "All attempts fail:\n#1: error value: \"panic func\"",
		},
		{
			name:    "panic with 2 times and delay with 1 second",
			fn:      pfn,
			tries:   2,
			want:    2,
			delay:   1000,
			wantStr: "All attempts fail:\n#1: error value: \"panic func\"",
		},
	}

	runWorker(botTest{
		args: &baseArgs{},
	}, func(b *xbot.Bot, tt botTest) {
		for _, tt := range tests {
			ts := time.Now()
			tried, err := b.RetryWhenPanic(tt.fn, tt.tries, tt.delay, 1)

			s.Equal(tt.want, tried, tt.name)

			if tt.wantStr == "" {
				s.Equal(tt.wantErr, err, tt.name)
			} else {
				s.NotNil(err, tt.name)
				s.Contains(err.Error(), tt.wantStr, tt.name)
			}

			cost := xutil.ElapsedSeconds(ts, 2)
			log.Debug().Float64("cost", cost).Msg(tt.name)

			s.GreaterOrEqual(cost*1000, cast.ToFloat64(tt.delay), tt.name)
		}
	})
}

func (s *botSuite) Test_05_BindPopovers() {
	// tow sites:
	// https://sandywalker.github.io/webui-popover/demo/#
	// https://www.jquery-az.com/boots/demo.php?ex=69.0_2
	// https://jobb.blocket.se/
	tests := []botTest{
		{
			name: "default",
			args: &baseArgs{},
		},
		{
			name: "bind with pops",
			args: &baseArgs{
				popovers: []string{"001", "div#main"},
			},
		},
	}

	for _, tt := range tests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			b.BindPopovers(tt.args.popovers)
			s.True(true)
		})
	}
}

var popTests = []botTest{
	{
		name: "with 0 popovers",
		args: &baseArgs{
			url:      blocket.url,
			popovers: []string{},
		},
		want: 0,
	},
	{
		name: "with non-exist pop",
		args: &baseArgs{
			url:      blocket.url,
			popovers: []string{blocket.nonExistSel},
		},
		want: 0,
	},
	{
		name: "jobb.blocket with 1 popovers",
		args: blocket,
		want: 1,
	},
	{
		name: "non clickable",
		args: &baseArgs{
			url:      blocket.url,
			popovers: []string{blocket.searchTerm},
		},
		want: 0,
	},
}

func (s *botSuite) Test_05_CloseIfHasPopovers() {
	// with popovers but without binding
	for _, test := range popTests {
		runWorker(test, func(b *xbot.Bot, tt botTest) {
			got := b.CloseIfHasPopovers()
			s.Equal(0, got, tt.name)
		})
	}

	for _, tt := range popTests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			b.BindPopovers(tt.args.popovers)
			if tt.want > 0 {
				pop := b.GetElem(tt.args.popovers[0])
				s.NotNil(pop, tt.args.popovers)
			}

			got := b.CloseIfHasPopovers()
			s.Equal(tt.want, got, tt.name)
		})
	}
}

func (s *botSuite) Test_05_ClickPopoverByEsc() {
	tests := []botTest{
		{
			name: "no selector",
			args: &baseArgs{
				url:      baidu.url,
				popovers: []string{},
			},
			want: 0,
		},
		{
			name: "blocket's popover",
			args: blocket,
			want: 0,
		},
	}

	for _, tt := range tests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			sel := ""

			if funk.NotEmpty(tt.args.popovers) {
				sel = tt.args.popovers[0]
				elem := b.GetElem(sel)
				s.NotNil(elem)
			}

			b.ClickPopoverByEsc(sel)

			if sel != "" {
				elem := b.GetElems(sel)
				s.Empty(elem, tt.name)
			}
		})
	}

}

func (s *botSuite) Test_06_PressEscape() {
	tests := []botTest{
		{
			name: "no selector",
			args: &baseArgs{
				url:      baidu.url,
				popovers: []string{},
			},
			want: 0,
		},
		{
			name: "blocket's popover",
			args: blocket,
			want: 0,
		},
	}

	for _, tt := range tests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			sel := ""

			if funk.NotEmpty(tt.args.popovers) {
				sel = tt.args.popovers[0]
				elem := b.GetElem(sel)
				s.NotNil(elem)
				b.MustPressEscape(sel)

				elem2 := b.GetElems(sel)
				s.Empty(elem2, tt.name)
			}
		})
	}
}

func (s *botSuite) Test_06_PressTab() {
	var tests = []botTest{
		{
			// don't run this too fast, or angel will trigger recaptcha
			name: "input ohio without press tab",
			args: &baseArgs{
				url:      "https://angel.co/location/united-states",
				location: "div[class*=locationWrapper] input[id*=react-select]",
				submit:   `button[type=submit]`,
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "non exist selector",
			args: &baseArgs{
				url:      baidu.url,
				location: blocket.nonExistSel,
				submit:   `button[type=submit]`,
			},
			wantErr: xbot.ErrorSelNotFound,
		},
	}

	// skip the first one, you should run it manually
	for _, tt := range tests[1:] {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			state := "ohio"
			_, e := b.FillBar(tt.args.location, state)
			if e != nil {
				s.Equal(tt.wantErr, e, tt.name)
				return
			}

			if tt.wantErr == nil {
				e := b.PressTab(tt.args.location)
				s.Nil(e)
			}
			b.MustScrollAndClick(tt.args.submit)
			err := b.EnsureUrlHas(state, xbot.BotTimeout(10))
			s.Equal(tt.wantErr, err, tt.name)
		})
	}
}

func (s *botSuite) Test_07_EnsureAnyElem() {
	tests := []botTest{
		{
			name:    "baidu default input",
			args:    baidu,
			sels:    append(baidu.has, baidu.no...),
			wantStr: baidu.has[0],
			wantErr: nil,
		},
		{
			name:    "blocket with popovers",
			args:    blocket,
			sels:    append(blocket.popovers, blocket.nonExistSel),
			wantStr: blocket.popovers[0],
			wantErr: nil,
		},
		{
			name: "",
			args: &baseArgs{
				url: baidu.url,
			},
			sels:    []string{baidu.nonExistSel, "span.hot-refresh-text@@@换一换"},
			wantStr: "span.hot-refresh-text@@@换一换",
			wantErr: nil,
		},
		{
			name: "context.deadlineExceededError",
			args: &baseArgs{
				url: baidu.url,
			},
			sels:    []string{baidu.nonExistSel, blocket.searchTerm},
			wantStr: "",
			wantErr: &rod.ErrTry{Value: context.DeadlineExceeded},
		},
	}

	handle := func(b *xbot.Bot, tt botTest) {
		got1, got2 := b.EnsureAnyElem(tt.sels...)
		s.Equal(tt.wantStr, got1, tt.name)
		if tt.wantErr != nil {
			s.ErrorIs(got2, tt.wantErr, tt.name)
		} else {
			s.Nil(tt.wantErr)
		}
	}

	handleMust := func(b *xbot.Bot, tt botTest) {
		if tt.wantErr == nil {
			got1 := b.MustEnsureAnyElem(tt.sels...)
			s.Equal(tt.wantStr, got1, tt.name)
		} else {
			s.Panics(func() {
				b.MustEnsureAnyElem(tt.sels...)
			}, tt.name)
		}
	}

	for _, tt := range tests {
		runWorker(tt, handle)
	}

	for _, tt := range tests {
		runWorker(tt, handleMust)
	}
}

func (s *botSuite) Test_08_EnsureUrlHas() {
	tests := []botTest{
		{
			name:    "contain empty search info in url",
			args:    blocket,
			wantErr: nil,
		},
		{
			name: "should not contain non exist",
			args: &baseArgs{
				url:    baidu.url,
				urlHas: "this_should_not_be_found",
				submit: baidu.submit,
			},
			wantErr: &rod.ErrTry{Value: context.DeadlineExceeded},
		},
	}

	for _, tt := range tests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			b.BindPopovers(tt.args.popovers)

			b.MustScrollAndClick(tt.args.submit)
			err := b.EnsureUrlHas(tt.args.urlHas)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr, tt.name)
			} else {
				s.Nil(tt.wantErr)
			}
		})
	}

}

func (s *botSuite) Test_09_MustEval() {
	tests := []botTest{
		{
			name:    "window height should be 728",
			args:    baidu,
			wantF64: 728,
		},
	}

	for _, tt := range tests {
		runWorker(tt, func(b *xbot.Bot, tt botTest) {
			script := `() => window.innerHeight`
			got := b.MustEval(script)

			s.Equal(tt.wantF64, cast.ToFloat64(got), tt.name)
		})
	}
}

func (s *botSuite) Test_10_FillBar() {
	tests := []botTest{
		{
			name:       "enter without submit",
			args:       baidu,
			wantStr:    "search google in baidu",
			withSubmit: false,
			wantErr:    nil,
		},
		{
			name:       "enter and submit",
			args:       baidu,
			wantStr:    "search google in baidu",
			withSubmit: true,
			wantErr:    nil,
		},
		{
			name: "with non-exist selector",
			args: &baseArgs{
				url:        baidu.url,
				searchTerm: baidu.nonExistSel,
			},
			wantStr:    "",
			withSubmit: true,
			wantErr:    xbot.ErrorSelNotFound,
		},
	}

	var handle = func(b *xbot.Bot, tt botTest) {
		ts := time.Now()
		got, err := b.FillBar(tt.args.searchTerm, tt.wantStr, xbot.InputSubmit(tt.withSubmit))
		elp := xutil.ElapsedSeconds(ts, 2)

		s.Equal(tt.wantStr, got, tt.name)
		c := float64(len(tt.wantStr) / 4)
		s.GreaterOrEqual(elp, c, tt.name)
		s.Equal(tt.wantErr, err, tt.name)
	}

	for _, tt := range tests {
		runWorker(tt, handle)
	}
}

var getElemData = []botTest{
	{
		name:    "empty selector: get 0",
		args:    &baseArgs{url: blocket.url, sel: ""},
		want:    0,
		wantInt: 0,
		to:      10,
		index:   0,
	},
	{
		name:    "plain txt selector: get 0",
		args:    &baseArgs{url: blocket.url, sel: "thisisaplaintext"},
		want:    0,
		wantInt: 0,
		to:      10,
		index:   0,
	},
	{
		name:    "non exist: get 0",
		args:    &baseArgs{url: blocket.url, sel: blocket.nonExistSel},
		want:    0,
		wantInt: 0,
		to:      10,
		index:   0,
	},
	{
		name:    "exist search: get 1",
		args:    &baseArgs{url: blocket.url, sel: blocket.searchTerm},
		want:    1,
		wantInt: 0,
		to:      5,
		index:   0,
		wantStr: "",
	},
	{
		name:    "exist and with index: get 1",
		args:    &baseArgs{sel: "div.checkbox>label>b"},
		want:    1,
		wantInt: 0,
		to:      5,
		index:   blocket.categoryIndex,
	},
	{
		name:    "exist and with index -1: get 1",
		args:    &baseArgs{sel: "div.checkbox>label>b"},
		want:    1,
		wantInt: 0,
		to:      5,
		index:   -1,
	},
	{
		name: "with @@@ by text, should find 1 in GetElem/0 in GetElems",
		args: &baseArgs{
			sel: blocket.category,
		},
		want:    0,
		wantInt: 1,
		to:      5,
		index:   0,
	},
	{
		name:    "with timeout 5 should find 1",
		args:    &baseArgs{sel: "div.checkbox>label>b"},
		want:    1,
		wantInt: 0,
		to:      5,
		index:   blocket.categoryIndex,
	},
}

func (s *botSuite) Test_1101_GetElemsAndNoDelay() {
	var handle = func(b *xbot.Bot, tt botTest) {
		ts := time.Now()
		got := b.GetElems(tt.args.sel)
		cost := xutil.ElapsedSeconds(ts, 2)

		if tt.index != 0 {
			s.LessOrEqual(tt.want, len(got), tt.name)
		} else {
			s.Equal(tt.want, len(got), tt.name)
		}
		s.LessOrEqual(cost, 1.0, tt.name)
	}

	var handleNoDelay = func(b *xbot.Bot, tt botTest) {
		ts := time.Now()
		got := b.GetElemWithoutDelay(tt.args.sel, tt.index)
		cost := xutil.ElapsedSeconds(ts, 2)

		if tt.want > 0 {
			s.NotNil(got, tt.name)
		} else {
			s.Nil(got, tt.name)
		}

		s.LessOrEqual(cost, 1.0, tt.name)
	}

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		for _, tt := range getElemData {
			handle(b, tt)
		}
		for _, tt := range getElemData {
			handleNoDelay(b, tt)
		}
	})
}

func (s *botSuite) Test_1102_GetElem_TestData() {
	var handle = func(b *xbot.Bot, tt botTest) {
		to := tt.to
		ts := time.Now()
		got := b.GetElem(tt.args.sel, xbot.ElemIndex(tt.index), xbot.BotTimeout(time.Duration(to)))

		cost := xutil.ElapsedSeconds(ts, 2)

		// elem
		if tt.want > 0 || tt.wantInt > 0 {
			s.NotNil(got, tt.name)
			s.LessOrEqual(cost, cast.ToFloat64(tt.to), tt.name)
		} else if tt.args.sel == "" {
			s.Nil(got, tt.name)
			s.LessOrEqual(cost, cast.ToFloat64(tt.to), tt.name)
		} else {
			s.Nil(got, tt.name)
			s.GreaterOrEqual(cost, cast.ToFloat64(tt.to), tt.name)
		}
	}

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		for _, tt := range getElemData {
			handle(b, tt)
		}
	})
}

var t1 = botTest{
	args: &baseArgs{
		url: blocket.url,
	},
}

func (s *botSuite) Test_1103_GetElem_RealWorld() {
	var handle = func(b *xbot.Bot, elem *rod.Element) {
		b.MustScrollAndClickElem(elem)

		selected := b.GetElem(blocket.confirmCategory, xbot.BotTimeout(0))
		s.Nil(selected, "confirmCategory take a few seconds to appear, check without delay, should get nil")

		selected = b.GetElem(blocket.confirmCategory)
		s.NotNil(selected, "this should appear finally")
		s.Equal("Data & IT", selected.MustText(), blocket.category)
		s.Contains(blocket.confirmCategory, selected.MustText(), blocket.category)
	}

	// test 1: we click blocket.category, and wait for blocket.confirmCategory to appear
	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		closePop(s, b)
		elem := b.GetElem(blocket.category)
		handle(b, elem)
	})
}

func (s *botSuite) Test_1201_GetElemAttr() {
	type tga struct {
		name   string
		args   baseArgs
		attr   string
		want   string
		noAttr bool
	}
	tests := []tga{
		{
			name:   "empty attr will get innerText",
			args:   baseArgs{sel: `div[id="recruitment-info"] h2[class="ui header head-text h1"]`},
			attr:   "",
			want:   "Letar du efter nästa stjärnkollega?",
			noAttr: false,
		},
		{
			name: "empty/innerText attr will get innerText, even if it is empty",
			args: baseArgs{
				sel: "a.menu-item>i.search",
			},
			attr:   "",
			want:   "",
			noAttr: true,
		},
		{
			name:   "default innerText",
			args:   baseArgs{sel: "a.menu-item>div.primary"},
			attr:   "innerText",
			want:   "Rekrytera",
			noAttr: false,
		},
		{
			name: "with attr",
			args: baseArgs{
				sel: blocket.searchTerm,
			},
			attr:   "maxlength",
			want:   "100",
			noAttr: false,
		},
		{
			name: "with attr",
			args: baseArgs{
				sel: blocket.searchTerm,
			},
			attr:   "value",
			want:   "",
			noAttr: false,
		},
		{
			name: "with non-exist attr",
			args: baseArgs{
				sel: blocket.searchTerm,
			},
			attr:   "nonExistSelValue",
			want:   "",
			noAttr: false,
		},
		{
			name: "with non-exist element",
			args: baseArgs{
				sel: blocket.nonExistSel,
			},
			attr:   "nonExistSelValue",
			want:   "",
			noAttr: false,
		},
	}

	var handle = func(b *xbot.Bot, tt tga) {
		var got string
		if tt.noAttr {
			got = b.GetElementAttr(tt.args.sel)
		} else {
			got = b.GetElementAttr(tt.args.sel, xbot.ElemAttr(tt.attr))
		}
		s.Equal(tt.want, got, tt.name)
	}

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		for _, tt := range tests {
			handle(b, tt)
		}
	})
}

func (s *botSuite) Test_1202_GetAttrOrProp() {
	runWorker(botTest{
		args: &baseArgs{url: blocket.url},
	}, func(b *xbot.Bot, tt botTest) {
		closePop(s, b)

		txt := "can you see"
		b.MustFillBar(blocket.searchTerm, txt)

		got1 := b.GetElementAttr(blocket.searchTerm)
		got2, err := b.GetElementProp(blocket.searchTerm)
		s.Nil(err)

		s.Equal(got1, txt)
		s.Equal(got1, got2, txt)
	})
}

func (s *botSuite) Test_13_GetWindowInnerHeight() {
	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		h := b.GetWindowInnerHeight()
		s.Equal(h, 728.0)
	}, false)
}

func (s *botSuite) Test_1401_ScrollAndClick_1() {
	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		v := b.ScrollAndClick("div.non-exist")
		s.NotNil(v, "div.non-exist")
	})

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		closePop(s, b)
		s.True(true, "scroll and close popovers")
	})

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		v := b.ScrollAndClick(blocket.submit)
		s.NotNil(v, "ScrollAndClick: category is covered by popover, should be false")
	})

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		sub := b.GetElem(blocket.submit)
		v := b.ScrollAndClickElem(sub, 2)
		s.NotNil(v, "ScrollAndClickElem: category is covered by popover, should be false")
	})

}

var closePop = func(s *botSuite, b *xbot.Bot) {
	pop := b.GetElemUntilInteractable(blocket.popovers[0])
	s.NotNil(pop)
	e := b.ClickElem(pop)
	s.Nil(e)
}

func (s *botSuite) Test_1402_ClickOrWithJs() {
	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		b.GetElemUntilInteractable(blocket.popovers[0])

		sub := b.GetElem(blocket.submit)
		s.NotNil(sub)
		e := b.ClickWithScript(sub)
		s.NotNil(e)
		s.ErrorIs(e, &rod.ErrNotInteractable{}, "click covered button")
	})

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		closePop(s, b)

		sub := b.GetElem(blocket.submit)
		s.NotNil(sub)
		e1 := b.ClickWithScript(sub)
		s.Nil(e1)

		b.Pg.MustWaitLoad()
		e2 := b.EnsureUrlHas(blocket.urlHas)
		s.Nil(e2)
	})
}

func (s *botSuite) Test_15_GetElemRSameAsGetElem() {
	// var url1, url2 string

	runWorker(t1, func(b *xbot.Bot, tt botTest) {
		closePop(s, b)
		elem2 := b.GetElem(blocket.category)
		b.MustClickElem(elem2)
		b.Pg.MustWaitLoad()
		b.MustEnsureUrlHas("data-it/")
	})

	// swg.Wait()
	// s.NotEmpty(url1)
	// s.Equal(url1, url2)
}
