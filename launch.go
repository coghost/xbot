package xbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/coghost/xutil"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
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
	loadProxy(l, cfg, opt)

	u, err := l.Launch()
	if err != nil {
		log.Fatal().Err(err).Msg("lauch failed")
	}
	return u
}

func loadProxyExtension(l *launcher.Launcher, cfg *BotConfig, opt BotOpts) {
	if cfg.ProxyLine != "" {
		dir := expandPath(cfg.ProxyRoot)
		extensionFolder, _, _ := xutil.NewChromeExtension(cfg.ProxyLine, dir)
		l.Set("load-extension", extensionFolder)
		log.Info().Str("extension_folder", extensionFolder).Msg("load proxy extension")
	}
}

func loadProxy(l *launcher.Launcher, cfg *BotConfig, opt BotOpts) {
	if v := cfg.ProxyServer; v != "" {
		l.Proxy(v)
		log.Info().Str("server", v).Msg("load proxy server")
	} else {
		loadProxyExtension(l, cfg, opt)
	}
}

func newUserModeLauncher(cfg *BotConfig, opt BotOpts) string {
	l := launcher.NewUserMode()

	loadProxy(l, cfg, opt)

	if b := cfg.Leakless; b {
		l.Leakless(b)
	}

	u, err := l.UserDataDir(cfg.UserDataDir).Launch()
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
	var page *rod.Page
	if cfg.WithStealth {
		log.Warn().Msg("running with stealth.js")
		page = stealth.MustPage(brw)
		go brw.EachEvent(func(e *proto.TargetTargetCreated) {
			if e.TargetInfo.Type != proto.TargetTargetInfoTypePage {
				return
			}
			brw.MustPageFromTargetID(e.TargetInfo.TargetID).MustEvalOnNewDocument(stealth.JS)
		})()
	} else {
		page = brw.MustPage("")
	}

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
