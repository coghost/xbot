package xbot

import (
	"os"
	"strings"

	"github.com/coghost/xutil"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
)

func NewBot(opts ...BotOptFunc) (bot *Bot) {
	opt := BotOpts{
		spawn:   true,
		panicBy: PanicByLogFatal,
		BotCfg:  defaultCfg,
	}
	BindBotOpts(&opt, opts...)

	if !opt.BotCfg.UserMode && opt.BotCfg.UserAgent == "" {
		panic(`UserAgent is required, please use xbot.BotUserAgent(ua) to bind it;
and you can visit https://www.whatismyip.com/user-agent/ to check your user-agent`)
	}

	bot = new(Bot)
	bot.Config = opt.BotCfg
	if opt.spawn {
		bot.Brw, bot.Pg = createBrwAndPage(opts...)
	}
	bot.SetTimeout()

	return bot
}

// NewDefaultBot creates a bot with default configs
func NewDefaultBot(spawn bool) *Bot {
	return NewBot(BotSpawn(spawn), BotUserAgent(UA))
}

func NewUserModeBot(opts ...BotOptFunc) (bot *Bot) {
	bc := NewDefaultBotCfg()
	bc.UserMode = true

	opt := BotOpts{BotCfg: bc}
	BindBotOpts(&opt, opts...)

	bot = NewBot(BotScreen(opt.BotCfg.Screen), WithBotConfig(opt.BotCfg))
	return bot
}

func Spawn(bot *Bot, opts ...BotOptFunc) {
	if bot.Brw == nil {
		bot.Brw, bot.Pg = createBrwAndPage(opts...)
	}
}

func createBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{BotCfg: defaultCfg}
	BindBotOpts(&opt, opts...)

	if opt.BotCfg.remoteServiceUrl != "" {
		return NewRemoteBrwAndPage(opts...)
	}

	if opt.BotCfg.UserMode {
		return NewUserModeBrwAndPage(opts...)
	}
	return NewBrwAndPage(opts...)
}

// NewBrwAndPage create and return a Browser and a blank page with window size 1366*768
func NewBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{
		BotCfg: defaultCfg,
	}
	BindBotOpts(&opt, opts...)
	cfg := opt.BotCfg

	u := newDefaultLanucher(cfg, opt)
	brw = customizeBrowser(u, cfg, opt)
	page = customizePage(brw, cfg, opt)

	return brw, page
}

// NewUserModeBrwAndPage run with user mode, will use system browser.
//
// we can integrate this with NewBrwAndPage, but there are too many if-else,
// so we just make a copy of NewBrwAndPage, and extract UserMode related logics
func NewUserModeBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{
		BotCfg: defaultCfg,
	}
	BindBotOpts(&opt, opts...)
	cfg := opt.BotCfg

	u := newUserModeLauncher(cfg, opt)
	brw = customizeBrowser(u, cfg, opt)
	if cfg.ClearCookies {
		log.Info().Msg("clear cookies in user-mode")
		brw.MustSetCookies()
	}
	page = customizePage(brw, cfg, opt)

	return brw, page
}

func NewRemoteBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{
		BotCfg: defaultCfg,
	}
	BindBotOpts(&opt, opts...)
	cfg := opt.BotCfg

	if strings.HasPrefix(cfg.remoteServiceUrl, "ws://") {
		l := newRemoteLauncher(opt)
		brw = rod.New().Client(l.MustClient()).MustConnect()
	} else {
		u := launcher.MustResolveURL(cfg.remoteServiceUrl)
		brw = rod.New().ControlURL(u).MustConnect()
	}
	page = customizePage(brw, cfg, opt)

	return brw, page
}

func setLauncher(l *launcher.Launcher, opt *BotOpts) *launcher.Launcher {
	l = l.
		Set("no-sandbox").
		Set("no-first-run").
		Set("no-startup-window").
		Set("disable-gpu").
		Set("disable-dev-shm-usage").
		Set("disable-web-security").
		Delete("use-mock-keychain").
		Set("disable-infobars").
		Set("enable-automation").
		Devtools(opt.BotCfg.Devtools).
		Headless(opt.BotCfg.Headless)
	return l
}

func bindUA(uaStr string) *proto.NetworkSetUserAgentOverride {
	log.Trace().Str("ua", uaStr).Msg("bind user-agent")
	ua := proto.NetworkSetUserAgentOverride{}
	ua.UserAgent = uaStr
	ua.AcceptLanguage = "en"
	return &ua
}

// expandPath will parse first `~` as user home dir path.
func expandPath(pathStr string) string {
	if len(pathStr) == 0 {
		return pathStr
	}

	if pathStr[0] != '~' {
		return pathStr
	}

	if len(pathStr) > 1 && pathStr[1] != '/' && pathStr[1] != '\\' {
		return pathStr
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return pathStr
	}

	return homeDir + pathStr[1:]
}

// ForceQuitBrowser will try to close browser by kill the process name
func ForceQuitBrowser(browserName string, opts ...BotOptFunc) error {
	opt := BotOpts{retry: 5}
	BindBotOpts(&opt, opts...)

	for i := 0; ; i++ {
		// pkill -a -i "Google Chrome"
		err := xutil.KillProcess(browserName)
		if err == nil {
			break
		}
		log.Error().Str("browser", browserName).Int("tried", i).Err(err).Msg("failed of close")
		if i >= opt.retry {
			return err
		}
		xutil.RandSleep(2.0, 2.5)
	}
	return nil
}
