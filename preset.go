package xbot

import (
	"errors"
)

const (
	SEP = "@@@"
)

const (
	LongTo   = 60
	MediumTo = 20
	ShortTo  = 5
	NapTo    = 2
)

const (
	PanicByDft = iota
	PanicByDump
	PanicByLogPanic
	PanicByLogError
)

var ErrorSelNotFound = errors.New("selector not found")
var defaultCfg = NewDefaultBotCfg()

func NewDefaultBotCfg() *BotConfig {
	return &BotConfig{
		Headless:       false,
		HighlightTimes: 1,
		ProxyRoot:      "/tmp/xbot/proxies",
		UserDataDir:    "/tmp/xbot/user_data",

		Screen:      1,
		Steps:       12,
		PageTimeout: 60,
		SlowMotion:  400,
		Width:       1366,
		Height:      728,

		PerInputLength: 7,
	}
}
