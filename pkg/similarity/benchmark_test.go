package similarity

import (
	"testing"

	"github.com/iamgideonidoko/signet/internal/models"
)

func BenchmarkExtractFeatures(b *testing.B) {
	calc := NewCalculator(DefaultWeights)
	signals := models.Signals{
		Canvas2DHash:        "abc123def456789",
		AudioHash:           "xyz789abc123def",
		WebGLVendor:         "NVIDIA Corporation",
		WebGLRenderer:       "GeForce GTX 1080/PCIe/SSE2",
		WebGLExtensions:     []string{"WEBGL_debug_renderer_info", "EXT_texture_filter_anisotropic"},
		HardwareConcurrency: 8,
		DeviceMemory:        16,
		ColorDepth:          24,
		MaxTouchPoints:      0,
		TimeZone:            "America/New_York",
		Languages:           []string{"en-US", "en"},
		Fonts:               []string{"Arial", "Helvetica", "Times New Roman"},
		ScreenWidth:         1920,
		ScreenHeight:        1080,
		Platform:            "MacIntel",
		UserAgent:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calc.ExtractFeatures(signals)
	}
}

func BenchmarkJaccardSimilarity(b *testing.B) {
	calc := NewCalculator(DefaultWeights)
	signals := models.Signals{
		Canvas2DHash:        "abc123def456789",
		AudioHash:           "xyz789abc123def",
		WebGLVendor:         "NVIDIA Corporation",
		WebGLRenderer:       "GeForce GTX 1080/PCIe/SSE2",
		HardwareConcurrency: 8,
		DeviceMemory:        16,
		TimeZone:            "America/New_York",
		Languages:           []string{"en-US", "en"},
	}

	v1 := calc.ExtractFeatures(signals)
	v2 := calc.ExtractFeatures(signals)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calc.JaccardSimilarity(v1, v2)
	}
}

func BenchmarkComputeHardwareHash(b *testing.B) {
	signals := models.Signals{
		Canvas2DHash:        "abc123def456789",
		AudioHash:           "xyz789abc123def",
		WebGLVendor:         "NVIDIA Corporation",
		WebGLRenderer:       "GeForce GTX 1080/PCIe/SSE2",
		HardwareConcurrency: 8,
		DeviceMemory:        16,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeHardwareHash(signals)
	}
}
