import type { IdentifyResponse, Signals } from "./types";

class SignetAgent {
  private readonly performanceStart: number;

  constructor() {
    this.performanceStart = performance.now();
  }

  async identify(apiEndpoint: string): Promise<IdentifyResponse> {
    const signals = await this.collectSignals();

    const elapsed = performance.now() - this.performanceStart;
    if (elapsed > 50) {
      console.warn(
        `[Signet] Collection took ${elapsed.toFixed(2)}ms (target: <50ms)`,
      );
    }

    const response = await fetch(apiEndpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ signals }),
    });

    if (!response.ok) {
      throw new Error(`Identification failed: ${response.statusText}`);
    }

    return response.json();
  }

  private async collectSignals(): Promise<Signals> {
    const [
      canvas2DHash,
      canvasWinding,
      webglData,
      audioHash,
      audioContextHash,
      fonts,
      plugins,
      mediaDevices,
      batteryPresent,
      permissionsHash,
      botDetection,
    ] = await Promise.all([
      this.getCanvas2DHash(),
      this.getCanvasWinding(),
      this.getWebGLData(),
      this.getAudioHash(),
      this.getAudioContextHash(),
      this.detectFonts(),
      this.getPlugins(),
      this.getMediaDevices(),
      this.checkBattery(),
      this.getPermissionsHash(),
      this.detectBot(),
    ]);

    return {
      // Canvas
      canvas_2d_hash: canvas2DHash,
      canvas_winding: canvasWinding,

      // WebGL
      webgl_vendor: webglData.vendor,
      webgl_renderer: webglData.renderer,
      webgl_extensions: webglData.extensions,
      webgl_params: webglData.params,
      webgl_hash: webglData.hash,

      // Audio
      audio_hash: audioHash,
      audio_context_hash: audioContextHash,

      // Hardware
      hardware_concurrency: navigator.hardwareConcurrency || 0,
      device_memory: (navigator as any).deviceMemory || 0,
      color_depth: screen.colorDepth || 0,
      pixel_ratio: window.devicePixelRatio || 1,
      max_touch_points: navigator.maxTouchPoints || 0,

      // Screen
      screen_width: screen.width,
      screen_height: screen.height,
      avail_width: screen.availWidth,
      avail_height: screen.availHeight,
      color_gamut: this.getColorGamut(),
      hdr_capable: this.checkHDR(),

      // System
      timezone: this.getTimeZone(),
      timezone_offset: new Date().getTimezoneOffset(),
      languages: navigator.languages ? Array.from(navigator.languages) : [],
      platform: navigator.platform,
      user_agent: navigator.userAgent,
      vendor: navigator.vendor,

      // Fonts
      fonts,

      // Bot Detection
      ...botDetection,

      // Advanced
      plugins,
      media_devices: mediaDevices,
      battery_present: batteryPresent,
      permissions_hash: permissionsHash,
      do_not_track: (navigator as any).doNotTrack,
    };
  }

  /**
   * Enhanced Canvas 2D fingerprinting
   */
  private async getCanvas2DHash(): Promise<string> {
    try {
      const canvas = document.createElement("canvas");
      canvas.width = 280;
      canvas.height = 60;
      const ctx = canvas.getContext("2d");

      if (!ctx) return "no_context";

      // Background gradient
      const gradient = ctx.createLinearGradient(0, 0, 280, 60);
      gradient.addColorStop(0, "#f00");
      gradient.addColorStop(0.5, "#0f0");
      gradient.addColorStop(1, "#00f");
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, 280, 60);

      // Text with multiple fonts and styles
      ctx.textBaseline = "alphabetic";
      ctx.fillStyle = "#f60";
      ctx.font = '11px "Arial"';
      ctx.fillText("Cwm fjordbank glyphs vext quiz, 😃", 2, 15);

      ctx.font = '13px "Times New Roman"';
      ctx.fillStyle = "rgba(102, 204, 0, 0.7)";
      ctx.fillText("Cwm fjordbank glyphs vext quiz, 😃", 4, 30);

      // Emoji rendering differences
      ctx.font = "16px sans-serif";
      ctx.fillText("🔐🌐💻🚀", 2, 50);

      // Subpixel rendering
      ctx.globalCompositeOperation = "multiply";
      ctx.fillStyle = "rgb(255,0,255)";
      ctx.beginPath();
      ctx.arc(50, 30, 20, 0, Math.PI * 2, true);
      ctx.fill();

      const imageData = ctx.getImageData(0, 0, 280, 60);
      return await this.hashPixelData(imageData.data);
    } catch {
      return "error";
    }
  }

  /**
   * Canvas winding test (detects certain spoofing)
   */
  private getCanvasWinding(): boolean {
    try {
      const canvas = document.createElement("canvas");
      const ctx = canvas.getContext("2d");
      if (!ctx) return false;

      ctx.rect(0, 0, 10, 10);
      ctx.rect(2, 2, 6, 6);
      return !ctx.isPointInPath(5, 5, "evenodd");
    } catch {
      return false;
    }
  }

  /**
   * Comprehensive WebGL fingerprinting
   */
  private getWebGLData(): {
    vendor: string;
    renderer: string;
    extensions: string[];
    params: Record<string, any>;
    hash: string;
  } {
    try {
      const canvas = document.createElement("canvas");
      const gl =
        canvas.getContext("webgl") ||
        (canvas.getContext(
          "experimental-webgl",
        ) as WebGLRenderingContext | null);

      if (!gl) {
        return {
          vendor: "none",
          renderer: "none",
          extensions: [],
          params: {},
          hash: "",
        };
      }

      // Get vendor and renderer
      const debugInfo = gl.getExtension("WEBGL_debug_renderer_info");
      const vendor = debugInfo
        ? gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL)
        : gl.getParameter(gl.VENDOR);
      const renderer = debugInfo
        ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL)
        : gl.getParameter(gl.RENDERER);

      const extensions = gl.getSupportedExtensions() || [];

      // Extract WebGL parameters
      const params: Record<string, any> = {};
      const paramNames = [
        "MAX_TEXTURE_SIZE",
        "MAX_VIEWPORT_DIMS",
        "MAX_VERTEX_ATTRIBS",
        "MAX_VERTEX_UNIFORM_VECTORS",
        "MAX_VARYING_VECTORS",
        "MAX_FRAGMENT_UNIFORM_VECTORS",
        "MAX_TEXTURE_IMAGE_UNITS",
        "MAX_RENDERBUFFER_SIZE",
        "MAX_COMBINED_TEXTURE_IMAGE_UNITS",
        "MAX_CUBE_MAP_TEXTURE_SIZE",
        "ALIASED_LINE_WIDTH_RANGE",
        "ALIASED_POINT_SIZE_RANGE",
        "SHADING_LANGUAGE_VERSION",
        "VERSION",
      ];

      for (const name of paramNames) {
        try {
          const value = gl.getParameter((gl as any)[name]);
          params[name] = Array.isArray(value) ? value.join(",") : String(value);
        } catch {}
      }

      // Generate WebGL hash from rendered scene
      const hash = this.renderWebGLScene(gl);

      return { vendor, renderer, extensions, params, hash };
    } catch {
      return {
        vendor: "error",
        renderer: "error",
        extensions: [],
        params: {},
        hash: "",
      };
    }
  }

  /**
   * Render WebGL scene for fingerprinting
   */
  private renderWebGLScene(gl: WebGLRenderingContext): string {
    try {
      const vertexShader = gl.createShader(gl.VERTEX_SHADER)!;
      gl.shaderSource(
        vertexShader,
        `
        attribute vec2 position;
        void main() { gl_Position = vec4(position, 0.0, 1.0); }
      `,
      );
      gl.compileShader(vertexShader);

      const fragmentShader = gl.createShader(gl.FRAGMENT_SHADER)!;
      gl.shaderSource(
        fragmentShader,
        `
        precision mediump float;
        void main() { gl_FragColor = vec4(1.0, 0.0, 0.5, 1.0); }
      `,
      );
      gl.compileShader(fragmentShader);

      const program = gl.createProgram()!;
      gl.attachShader(program, vertexShader);
      gl.attachShader(program, fragmentShader);
      gl.linkProgram(program);
      gl.useProgram(program);

      // Draw triangle
      const vertices = new Float32Array([-0.5, -0.5, 0.5, -0.5, 0.0, 0.5]);
      const buffer = gl.createBuffer();
      gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
      gl.bufferData(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW);

      const position = gl.getAttribLocation(program, "position");
      gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0);
      gl.enableVertexAttribArray(position);

      gl.drawArrays(gl.TRIANGLES, 0, 3);

      // Read pixels
      const pixels = new Uint8Array(4);
      gl.readPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, pixels);

      return Array.from(pixels).join(",");
    } catch {
      return "";
    }
  }

  /**
   * Enhanced AudioContext fingerprinting
   */
  private async getAudioContextHash(): Promise<string> {
    try {
      const AudioContext =
        (window as any).AudioContext || (window as any).webkitAudioContext;
      if (!AudioContext) return "no_audio_context";

      const context = new AudioContext();
      const params: string[] = [
        `sr:${context.sampleRate}`,
        `ch:${context.destination.channelCount}`,
        `state:${context.state}`,
      ];

      context.close();
      return params.join("|");
    } catch {
      return "error";
    }
  }

  private async getAudioHash(): Promise<string> {
    try {
      const AudioContext =
        (window as any).AudioContext || (window as any).webkitAudioContext;
      if (!AudioContext) return "no_audio_context";

      const context = new AudioContext();
      const oscillator = context.createOscillator();
      const compressor = context.createDynamicsCompressor();
      const analyser = context.createAnalyser();

      oscillator.type = "triangle";
      oscillator.frequency.value = 10000;

      compressor.threshold.value = -50;
      compressor.knee.value = 40;
      compressor.ratio.value = 12;
      compressor.attack.value = 0;
      compressor.release.value = 0.25;

      oscillator.connect(compressor);
      compressor.connect(analyser);
      compressor.connect(context.destination);

      oscillator.start(0);

      await new Promise((resolve) => setTimeout(resolve, 10));

      const buffer = new Float32Array(analyser.frequencyBinCount);
      analyser.getFloatFrequencyData(buffer);

      oscillator.stop();
      oscillator.disconnect();
      compressor.disconnect();
      context.close();

      return await this.hashFloatArray(buffer);
    } catch {
      return "error";
    }
  }

  /**
   * Extended font detection
   */
  private async detectFonts(): Promise<string[]> {
    const baseFonts = ["monospace", "sans-serif", "serif"];
    const testFonts = [
      "Arial",
      "Verdana",
      "Times New Roman",
      "Courier New",
      "Georgia",
      "Palatino",
      "Garamond",
      "Bookman",
      "Comic Sans MS",
      "Trebuchet MS",
      "Impact",
      "Helvetica",
      "Tahoma",
      "Geneva",
      "Consolas",
      "Calibri",
      "Cambria",
      "Century Gothic",
      "Franklin Gothic",
      "Lucida Console",
      "Monaco",
      "Courier",
      "Menlo",
      "Ubuntu",
      "Roboto",
      "Open Sans",
      "Segoe UI",
      "Avenir",
      "Futura",
    ];

    const detectedFonts: string[] = [];
    const testString = "mmmmmmmmmmlli";
    const testSize = "72px";

    const span = document.createElement("span");
    span.style.position = "absolute";
    span.style.left = "-9999px";
    span.style.fontSize = testSize;
    span.textContent = testString;
    document.body.appendChild(span);

    const baseWidths: { [key: string]: number } = {};
    for (const baseFont of baseFonts) {
      span.style.fontFamily = baseFont;
      baseWidths[baseFont] = span.offsetWidth;
    }

    for (const testFont of testFonts) {
      let detected = false;
      for (const baseFont of baseFonts) {
        span.style.fontFamily = `'${testFont}', ${baseFont}`;
        if (
          span.offsetWidth !== baseWidths[baseFont] ||
          span.offsetHeight !== 72
        ) {
          detected = true;
          break;
        }
      }
      if (detected) {
        detectedFonts.push(testFont);
      }
    }

    document.body.removeChild(span);
    return detectedFonts;
  }

  /**
   * Comprehensive bot detection
   */
  private async detectBot(): Promise<{
    webdriver: boolean;
    chrome_present: boolean;
    phantom_present: boolean;
    headless_chrome: boolean;
    selenium_present: boolean;
    automation_present: boolean;
  }> {
    const win = window as any;

    return {
      webdriver: !!navigator.webdriver,
      chrome_present: !!win.chrome,
      phantom_present: !!win.callPhantom || !!win._phantom,
      headless_chrome: /HeadlessChrome/.test(navigator.userAgent),
      selenium_present: !!win.document.$cdc_ || !!win.document.$wdc_,
      automation_present:
        !!navigator.webdriver ||
        !!win.__webdriver_script_fn ||
        !!win.domAutomation ||
        !!win.domAutomationController,
    };
  }

  /**
   * Get installed plugins
   */
  private async getPlugins(): Promise<string[]> {
    try {
      const plugins = Array.from(navigator.plugins || [])
        .map((p) => p.name)
        .slice(0, 10);
      return plugins;
    } catch {
      return [];
    }
  }

  /**
   * Get media devices count
   */
  private async getMediaDevices(): Promise<number> {
    try {
      if (!navigator.mediaDevices?.enumerateDevices) return 0;
      const devices = await navigator.mediaDevices.enumerateDevices();
      return devices.length;
    } catch {
      return 0;
    }
  }

  /**
   * Check battery API presence
   */
  private async checkBattery(): Promise<boolean> {
    try {
      if (!(navigator as any).getBattery) return false;
      await (navigator as any).getBattery();
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get permissions hash
   */
  private async getPermissionsHash(): Promise<string> {
    try {
      const permissions = [
        "notifications",
        "geolocation",
        "camera",
        "microphone",
      ];
      const states: string[] = [];

      for (const perm of permissions) {
        try {
          const result = await navigator.permissions.query({
            name: perm as any,
          });
          states.push(`${perm}:${result.state}`);
        } catch {}
      }

      return states.join("|");
    } catch {
      return "";
    }
  }

  private getColorGamut(): string {
    if (window.matchMedia("(color-gamut: p3)").matches) return "p3";
    if (window.matchMedia("(color-gamut: srgb)").matches) return "srgb";
    if (window.matchMedia("(color-gamut: rec2020)").matches) return "rec2020";
    return "unknown";
  }

  private checkHDR(): boolean {
    return window.matchMedia("(dynamic-range: high)").matches;
  }

  private getTimeZone(): string {
    try {
      return Intl.DateTimeFormat().resolvedOptions().timeZone;
    } catch {
      return "unknown";
    }
  }

  private async hashPixelData(data: Uint8ClampedArray): Promise<string> {
    return this.hashData(data);
  }

  private async hashFloatArray(data: Float32Array): Promise<string> {
    const buffer = new Uint8Array(data.buffer);
    return this.hashData(buffer);
  }

  private async hashData(
    data: Uint8Array | Uint8ClampedArray,
  ): Promise<string> {
    try {
      if (crypto.subtle) {
        const hashBuffer = await crypto.subtle.digest("SHA-256", data);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
      }
    } catch {}

    return this.simpleHash(data);
  }

  private simpleHash(data: Uint8Array | Uint8ClampedArray): string {
    let hash = 0;
    for (let i = 0; i < Math.min(data.length, 1000); i++) {
      hash = (hash << 5) - hash + data[i];
      hash = hash & hash;
    }
    return Math.abs(hash).toString(16);
  }
}

const Signet = new SignetAgent();
export default Signet;
