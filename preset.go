package xbot

import (
	"errors"
)

const (
	SEP = "@@@"
)

const clickButtonTimes = 1

const (
	LongTo   = 60
	MediumTo = 20
	ShortTo  = 5
	NapTo    = 2
	ZeroTo   = 0

	NearlyNonTo = 0.1
)

const UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"

type BotPanicType int

const (
	PanicByDft BotPanicType = iota
	PanicByDump
	PanicByLogError
	PanicByLogFatal
	PanicByLogPanic
)

var ErrorSelNotFound = errors.New("selector not found")
var defaultCfg = NewDefaultBotCfg()

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
