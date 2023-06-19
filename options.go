package xbot

import "github.com/go-rod/rod"

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
	Timeout int

	// proxyLine with format `host:port:username:password:<OTHER>`
	// proxyLine string

	BotCfg *BotConfig

	sleepSecBeforeAction float64
	scrollAsHuman        bool

	retry int

	root *rod.Element
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

func WithDefaultBotConfig() BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg = defaultCfg
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

// BotProxyLine is with format `host:port:username:password:<OTHER>`
func BotProxyLine(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.ProxyLine = s
	}
}

// BotProxyServer is with format `host:port`
//
// proxy server has higher priority than proxy-line,
// so if both proxy-server and proxy-line provided, will use proxy-server
//
//	e.g. BotProxyServer("127.0.0.1:12345")
func BotProxyServer(s string) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.ProxyServer = s
	}
}

// BotScreen is the x position of screen,
// you can set it to value less than 0 if you have multiple screens
//
//	e.g. you have a second display, with resolution 2560*1440, then you can set BotScreen(-2560)
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

func BotTimeout(i int) BotOptFunc {
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

func WithStealth(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.WithStealth = b
	}
}

// WithLeakless works only with user_mode
func WithLeakless(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.Leakless = b
	}
}

// WithLeakless works only with user_mode
func DisableCookies(b bool) BotOptFunc {
	return func(o *BotOpts) {
		o.BotCfg.ClearCookies = b
	}
}

func WithRoot(root *rod.Element) BotOptFunc {
	return func(o *BotOpts) {
		o.root = root
	}
}
