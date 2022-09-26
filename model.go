package xbot

import (
	"time"

	"github.com/go-rod/rod"
)

type Box struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

type Bot struct {
	WithHighlight bool

	// Steps: means take how many steps to scroll to position
	Steps int

	LongTo   time.Duration
	MediumTo time.Duration
	ShortTo  time.Duration
	NapTo    time.Duration

	popovers []string

	Brw *rod.Browser
	Pg  *rod.Page

	// selector
	selector interface{}

	Config *BotConfig
}

// BotConfig is used to config bot options, which is usually read from config file
type BotConfig struct {
	BinFile     string `ini:"bin_file"`
	UserAgent   string `ini:"user_agent"`
	UserMode    bool   `ini:"user_mode"`
	Maximize    bool   `ini:"maximize"`
	UserDataDir string `ini:"user_data_dir"`
	Highlight   bool   `ini:"highlight"`
	Screen      int    `ini:"screen"`
	Steps       int    `ini:"steps"`

	// when we want to inputAsHuman, each time input how many chars
	PerInputLength int `ini:"per_input_length"`

	// height, width
	Width  int `ini:"width"`
	Height int `ini:"height"`

	ViewOffsetWidth  int `ini:"view_offset_width"`
	ViewOffsetHeight int `ini:"view_offset_height"`

	NoDefaultDevice bool `ini:"no_default_device"`

	// ProxyRoot automatically created proxy saving path
	ProxyRoot string `ini:"proxy_root"`
	ProxyLine string `ini:"proxy_line"`

	PageTimeout int `ini:"page_timeout"`

	// .rod->show
	Headless bool `ini:"headless"`
	// .rod->slow
	SlowMotion int `ini:"slow_motion"`
	// .rod->trace
	Trace bool `ini:"trace"`
	// Incognito mode
	Incognito bool `ini:"incognito"`

	Devtools bool `ini:"devtools"`

	//
	HighlightTimes int `ini:"highlight_times"`

	ScrollDistanceBase float64 `ini:"scroll_distance_base"`

	OffsetToTop float64 `ini:"offset_to_top"`

	// auto_recaptcha
	AutoRecaptcha bool `ini:"auto_recaptcha"`
}
