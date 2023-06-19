package xbot

import (
	"errors"
)

const (
	SEP = "@@@"
)

const clickButtonTimes = 1

const (
	// ZeroToSec no timeout in second
	ZeroToSec = 0
	// NapToSec a nap timeout in second
	NapToSec = 2
	// ShortToSec short timeout in second
	ShortToSec = 5
	// MediumToSec medium timeout in second
	MediumToSec = 20
	// LongToSec long timeout in second
	LongToSec = 60
	// NearlyNonToSec a very short timeout in second
	NearlyNonToSec = 0.1
)

const UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"

const (
	BrowserChrome = "Google Chrome"
)

type BotPanicType int

const (
	PanicByDft BotPanicType = iota
	PanicByDump
	PanicByLogError
	PanicByLogFatal
	PanicByLogPanic
)

var (
	ErrorSelNotFound = errors.New("selector not found")
	defaultCfg       = NewDefaultBotCfg()
)

func NewDefaultBotCfg() *BotConfig {
	return &BotConfig{
		Headless:       false,
		Highlight:      true,
		HighlightTimes: 1,
		ProxyRoot:      "/tmp/xbot/proxies",
		UserDataDir:    "/tmp/xbot/user_data",

		Screen:      1,
		Steps:       12,
		PageTimeout: 60,
		SlowMotion:  400,
		Width:       1366,
		Height:      728,

		NoDefaultDevice: true,
		Incognito:       false,

		PerInputLength: 7,
	}
}
