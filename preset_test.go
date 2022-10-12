package xbot_test

import (
	"testing"

	"github.com/coghost/xbot"
	"github.com/stretchr/testify/suite"
)

type PresetSuite struct {
	suite.Suite
}

func TestPreset(t *testing.T) {
	suite.Run(t, new(PresetSuite))
}

func (s *PresetSuite) SetupSuite() {
}

func (s *PresetSuite) TearDownSuite() {
}

func (s *PresetSuite) Test_01_new() {
	cfg := xbot.NewDefaultBotCfg()
	s.Equal(cfg.AutoRecaptcha, false)

	cf1 := &xbot.BotConfig{
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

	s.Equal(cf1, cfg)
}
