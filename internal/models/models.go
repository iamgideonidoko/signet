package models

type Signals struct {
	// gpu /rendering
	Canvas2DHash    string         `json:"canvas_2d_hash"`
	CanvasWinding   bool           `json:"canvas_winding"`
	WebGLVendor     string         `json:"webgl_vendor"`
	WebGLRenderer   string         `json:"webgl_renderer"`
	WebGLExtensions []string       `json:"webgl_extensions"`
	WebGLParams     map[string]any `json:"webgl_params"`
	WebGLHash       string         `json:"webgl_hash"`

	// hardware dynamics
	AudioHash           string  `json:"audio_hash"`
	AudioContextHash    string  `json:"audio_context_hash"`
	HardwareConcurrency int     `json:"hardware_concurrency"`
	DeviceMemory        float64 `json:"device_memory"`
	ColorDepth          int     `json:"color_depth"`
	PixelRatio          float64 `json:"pixel_ratio"`
	MaxTouchPoints      int     `json:"max_touch_points"`

	// screen / display
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
	AvailWidth   int    `json:"avail_width"`
	AvailHeight  int    `json:"avail_height"`
	ColorGamut   string `json:"color_gamut"`
	HDRCapable   bool   `json:"hdr_capable"`

	// system environment
	TimeZone       string   `json:"timezone"`
	TimezoneOffset int      `json:"timezone_offset"`
	Languages      []string `json:"languages"`
	Platform       string   `json:"platform"`
	UserAgent      string   `json:"user_agent"`
	Vendor         string   `json:"vendor"`
	Fonts          []string `json:"fonts"`

	// bot detection
	WebDriver         bool `json:"webdriver"`
	ChromePresent     bool `json:"chrome_present"`
	PhantomPresent    bool `json:"phantom_present"`
	HeadlessChrome    bool `json:"headless_chrome"`
	SeleniumPresent   bool `json:"selenium_present"`
	AutomationPresent bool `json:"automation_present"`

	// advanced
	Plugins         []string `json:"plugins"`
	MediaDevices    int      `json:"media_devices"`
	BatteryPresent  bool     `json:"battery_present"`
	PermissionsHash string   `json:"permissions_hash"`
	DoNotTrack      string   `json:"do_not_track,omitempty"`
}
