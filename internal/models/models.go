package models

import (
	"time"

	"github.com/google/uuid"
)

// Visitor represents a unique browser/device identity.
type Visitor struct {
	VisitorID   uuid.UUID `json:"visitor_id" db:"visitor_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	TrustScore  float64   `json:"trust_score" db:"trust_score"`
	FirstSeenIP *string   `json:"first_seen_ip,omitempty" db:"first_seen_ip"`
	LastSeenIP  *string   `json:"last_seen_ip,omitempty" db:"last_seen_ip"`
	VisitCount  int       `json:"visit_count" db:"visit_count"`
}

// Identification represents a single fingerprint submission.
type Identification struct {
	RequestID       uuid.UUID `json:"request_id" db:"request_id"`
	VisitorID       uuid.UUID `json:"visitor_id" db:"visitor_id"`
	IPAddress       string    `json:"ip_address" db:"ip_address"`
	UserAgent       *string   `json:"user_agent,omitempty" db:"user_agent"`
	Signals         Signals   `json:"signals" db:"signals"`
	ConfidenceScore float64   `json:"confidence_score" db:"confidence_score"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	HardwareHash    string    `json:"hardware_hash" db:"hardware_hash"`
	IsBot           bool      `json:"is_bot" db:"is_bot"`
}

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

// IdentifyRequest is the incoming fingerprint payload.
type IdentifyRequest struct {
	Signals   Signals `json:"signals" validate:"required"`
	IPAddress string  `json:"-"` // Populated from request context
}

// IdentifyResponse is returned to the client.
type IdentifyResponse struct {
	VisitorID  uuid.UUID `json:"visitor_id"`
	Confidence float64   `json:"confidence"`
	IsNew      bool      `json:"is_new"`
	RequestID  uuid.UUID `json:"request_id"`
}

// VisitorAnalytics represents aggregated metrics.
type VisitorAnalytics struct {
	Date           string  `json:"date" db:"date"`
	UniqueVisitors int     `json:"unique_visitors" db:"unique_visitors"`
	TotalRequests  int     `json:"total_requests" db:"total_requests"`
	AvgConfidence  float64 `json:"avg_confidence" db:"avg_confidence"`
	BotRequests    int     `json:"bot_requests" db:"bot_requests"`
}
