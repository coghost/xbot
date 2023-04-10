package xbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/coghost/xutil"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/rs/zerolog/log"
)

func newDefaultLanucher(cfg *BotConfig, opt BotOpts) string {
	var l *launcher.Launcher
	l = launcher.New()
	if cfg.BinFile != "" {
		l = l.Bin(cfg.BinFile)
	}

	// only default launcher require this
	l = setLauncher(l, &opt)

	if opt.proxyLine != "" {
		dir := expandPath(cfg.ProxyRoot)
		extensionFolder, _, _ := xutil.NewChromeExtension(opt.proxyLine, dir)
		l.Set("load-extension", extensionFolder)
		log.Debug().Str("extension_folder", extensionFolder).Msg("load proxy extension")
	}

	u, err := l.Launch()
	if err != nil {
		log.Fatal().Err(err).Msg("lauch failed")
	}
	return u
}

func newUserModeLauncher(opt BotOpts) string {
	u, err := launcher.NewUserMode().Launch()
	if err != nil {
		s := fmt.Sprintf("%s", err)
		if strings.Contains(s, "[launcher] Failed to get the debug url: Opening in existing browser session") {
			fmt.Printf("%[1]s\nlaunch chrome browser failed, please make sure it is closed, and then run again\n%[1]s\n", strings.Repeat("=", 32))
			log.Fatal().Err(err).Msg("")
		} else {
			log.Fatal().Err(err).Msg("cannot launch browser")
		}
	}
	return u
}

func newRemoteLauncher(opt BotOpts) *launcher.Launcher {
	l := launcher.MustNewManaged(opt.BotCfg.remoteServiceUrl)
	l.Set("disable-gpu")
	// Launch with headful mode
	l.Headless(false).XVFB("--server-num=5", "--server-args=-screen 0 1600x900x16")
	return l
}

func customizeBrowser(u string, cfg *BotConfig, opt BotOpts) *rod.Browser {
	browser := rod.New().ControlURL(u).MustConnect()

	if cfg.NoDefaultDevice {
		browser = browser.NoDefaultDevice()
	}
	if cfg.Incognito {
		browser = browser.MustIncognito()
	}

	slow := xutil.AorB(cfg.SlowMotion, 500)
	browser.SlowMotion(time.Millisecond * time.Duration(slow))
	browser.Trace(cfg.Trace)

	return browser
}

func customizePage(brw *rod.Browser, cfg *BotConfig, opt BotOpts) *rod.Page {
	page := brw.MustPage("")

	if opt.BotCfg.UserAgent != "" {
		ua := bindUA(opt.BotCfg.UserAgent)
		page = page.MustSetUserAgent(ua)
	}

	if cfg.Maximize {
		page.MustWindowMaximize()
		return page
	}

	w, h := cfg.Width, cfg.Height
	vw := xutil.AorB(cfg.ViewOffsetWidth, 0)
	vh := xutil.AorB(cfg.ViewOffsetHeight, 0)
	page = page.MustSetWindow(opt.BotCfg.Screen, 0, w, h)
	if vw != 0 || vh != 0 {
		page = page.MustSetViewport(w-vw, h-vh, 0.0, false)
	}

	return page
}
