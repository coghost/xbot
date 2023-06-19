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
	panicBy BotPanicType

	longToSec   time.Duration
	mediumToSec time.Duration
	shortToSec  time.Duration
	NapToSec    time.Duration

	popovers []string

	Brw *rod.Browser
	Pg  *rod.Page

	Iframe   *rod.Page
	PrevPage *rod.Page

	// selector
	selector interface{}

	root *rod.Element

	Config *BotConfig

	ScrollAsHuman *ScrollAsHuman
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

	ProxyServer string `ini:"proxy_server"`
	Leakless    bool   `ini:"leakless"`

	PageTimeout int `ini:"page_timeout"`

	// .rod->show
	Headless bool `ini:"headless"`
	// .rod->slow in milliseconds
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

	// remoteServiceUrl
	// two types:
	//  - ws://ip:port this will launch remote browser
	//  - ip:port this only connect to remote browser
	remoteServiceUrl string `ini:"remote_service_url"`

	WithStealth bool `ini:"with_stealth"`

	ClearCookies bool `ini:"clear_cookies"`
}

type ScrollAsHuman struct {
	enabled          bool
	longSleepChance  float64
	shortSleepChance float64
	scrollUpChance   float64
}
