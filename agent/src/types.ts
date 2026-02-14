export interface Signals {
  // Canvas & Graphics
  canvas_2d_hash: string;
  canvas_winding: boolean;
  webgl_vendor: string;
  webgl_renderer: string;
  webgl_extensions: string[];
  webgl_params: Record<string, any>;
  webgl_hash: string;

  // Audio
  audio_hash: string;
  audio_context_hash: string;

  // Hardware
  hardware_concurrency: number;
  device_memory: number;
  color_depth: number;
  pixel_ratio: number;
  max_touch_points: number;

  // Screen & Display
  screen_width: number;
  screen_height: number;
  avail_width: number;
  avail_height: number;
  color_gamut: string;
  hdr_capable: boolean;

  // System
  timezone: string;
  timezone_offset: number;
  languages: string[];
  platform: string;
  user_agent: string;
  vendor: string;

  // Fonts
  fonts: string[];

  // Bot Detection
  webdriver: boolean;
  chrome_present: boolean;
  phantom_present: boolean;
  headless_chrome: boolean;
  selenium_present: boolean;
  automation_present: boolean;

  // Advanced
  plugins: string[];
  media_devices: number;
  battery_present: boolean;
  permissions_hash: string;
  do_not_track?: string;
}

export interface IdentifyResponse {
  visitor_id: string;
  confidence: number;
  is_new: boolean;
  request_id: string;
}
