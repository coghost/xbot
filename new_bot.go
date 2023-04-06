package xbot

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coghost/xutil"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
)

func NewBot(opts ...BotOptFunc) (bot *Bot) {
	opt := BotOpts{
		spawn:     true,
		Screen:    0,
		Headless:  false,
		Highlight: true,
		Steps:     16,
		BotCfg:    defaultCfg,
	}
	BindBotOpts(&opt, opts...)

	if !opt.BotCfg.UserMode && opt.UserAgent == "" {
		panic(`UserAgent is required, please use xbot.BotUserAgent(ua) to bind it;
and you can visit https://www.whatismyip.com/user-agent/ to check your user-agent`)
	}

	bot = new(Bot)
	bot.Config = opt.BotCfg
	if opt.spawn {
		bot.Brw, bot.Pg = createBrwAndPage(opts...)
	}
	bot.SetTimeout()
	bot.WithHighlight = opt.Highlight
	bot.Steps = opt.Steps

	return bot
}

// NewDefaultBot creates a bot with default configs
func NewDefaultBot(spawn bool) *Bot {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"
	return NewBot(BotSpawn(spawn), BotUserAgent(ua))
}

func NewUserModeBot(opts ...BotOptFunc) (bot *Bot) {
	bc := NewDefaultBotCfg()
	bc.UserMode = true

	opt := BotOpts{Screen: 1, BotCfg: bc}
	BindBotOpts(&opt, opts...)

	bot = NewBot(BotScreen(opt.Screen), WithBotConfig(opt.BotCfg))
	return bot
}

func Spawn(bot *Bot, opts ...BotOptFunc) {
	bot.Brw, bot.Pg = createBrwAndPage(opts...)
}

func createBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{BotCfg: defaultCfg}
	BindBotOpts(&opt, opts...)

	if opt.BotCfg.UserMode {
		return NewUserModeBrwAndPage(opts...)
	}
	return NewBrwAndPage(opts...)
}

// NewBrwAndPage create and return an Browser and a blank page with window size 1366*768
func NewBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{
		Screen:   0,
		Headless: false,
		BotCfg:   defaultCfg,
	}
	BindBotOpts(&opt, opts...)

	cfg := opt.BotCfg

	var l *launcher.Launcher
	l = launcher.New()
	if cfg.BinFile != "" {
		l = l.Bin(cfg.BinFile)
	}

	l = setLauncher(l, &opt)

	if opt.ProxyLine != "" {
		dir := expandPath(cfg.ProxyRoot)
		extensionFolder, _, _ := xutil.NewChromeExtension(opt.ProxyLine, dir)
		l.Set("load-extension", extensionFolder)
		log.Debug().Str("extension_folder", extensionFolder).Msg("load proxy extension")
	}

	u := l.MustLaunch()
	brw = rod.New().ControlURL(u).MustConnect()

	if cfg.NoDefaultDevice {
		brw = brw.NoDefaultDevice()
	}
	if cfg.Incognito {
		brw = brw.MustIncognito()
	}

	slow := xutil.AorB(cfg.SlowMotion, 500)
	brw.SlowMotion(time.Millisecond * time.Duration(slow))
	brw.Trace(cfg.Trace)

	page = brw.MustPage("")

	w, h := cfg.Width, cfg.Height
	vw := xutil.AorB(cfg.ViewOffsetWidth, 0)
	vh := xutil.AorB(cfg.ViewOffsetHeight, 0)
	ua := bindUA(opt.UserAgent)
	page.MustSetUserAgent(ua).MustSetWindow(opt.Screen, 0, w, h).MustSetViewport(w-vw, h-vh, 0.0, false)

	if cfg.Maximize {
		page.MustWindowMaximize()
	}
	return
}

// NewUserModeBrwAndPage run with user mode, will use system browser.
//
// we can integrate this with NewBrwAndPage, but there're too many if-else
// so we just make a copy of NewBrwAndPage, and extract UserMode related logics
func NewUserModeBrwAndPage(opts ...BotOptFunc) (brw *rod.Browser, page *rod.Page) {
	opt := BotOpts{
		Screen:   0,
		Headless: false,
		BotCfg:   defaultCfg,
	}
	BindBotOpts(&opt, opts...)
	cfg := opt.BotCfg

	u, err := launcher.NewUserMode().Launch()
	if err != nil {
		estr := fmt.Sprintf("%s", err)
		if strings.Contains(estr, "[launcher] Failed to get the debug url: Opening in existing browser session") {
			fmt.Printf("%[1]s\nlaunch chrome browser failed, please make sure it is closed, and then run again\n%[1]s\n", strings.Repeat("=", 32))
			log.Fatal().Err(err).Msg("")
		} else {
			log.Fatal().Err(err).Msg("cannot launch browser")
		}
	}

	brw = rod.New().ControlURL(u).MustConnect().NoDefaultDevice()

	slow := xutil.AorB(cfg.SlowMotion, 500)
	brw.SlowMotion(time.Millisecond * time.Duration(slow))
	brw.Trace(cfg.Trace)

	page = brw.MustPage("")

	if cfg.Maximize {
		page.MustWindowMaximize()
		return
	}

	w, h := cfg.Width, cfg.Height
	vw := xutil.AorB(cfg.ViewOffsetWidth, 0)
	vh := xutil.AorB(cfg.ViewOffsetHeight, 0)

	page.MustSetWindow(opt.Screen, 0, w, h)
	if vw != 0 || vh != 0 {
		page.MustSetViewport(w-vw, h-vh, 0.0, false)
	}
	return
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
		Headless(opt.Headless)
	return l
}

func bindUA(uaStr string) *proto.NetworkSetUserAgentOverride {
	log.Debug().Str("ua", uaStr).Msg("bind user-agent")
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
