package similarity

import (
	"testing"

	"github.com/iamgideonidoko/signet/internal/models"
)

func TestJaccardSimilarity_IdenticalSignals(t *testing.T) {
	calc := NewCalculator(DefaultWeights)

	signals := models.Signals{
		Canvas2DHash:        "abc123",
		AudioHash:           "def456",
		WebGLVendor:         "NVIDIA",
		WebGLRenderer:       "GeForce GTX 1080",
		HardwareConcurrency: 8,
		DeviceMemory:        16,
		TimeZone:            "America/New_York",
		Languages:           []string{"en-US", "en"},
	}

	v1 := calc.ExtractFeatures(signals)
	v2 := calc.ExtractFeatures(signals)

	similarity := calc.JaccardSimilarity(v1, v2)

	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical signals, got %.2f", similarity)
	}
}

func TestJaccardSimilarity_SoftwareChange(t *testing.T) {
	calc := NewCalculator(DefaultWeights)

	base := models.Signals{
		Canvas2DHash:        "abc123",
		AudioHash:           "def456",
		WebGLVendor:         "NVIDIA",
		WebGLRenderer:       "GeForce GTX 1080",
		HardwareConcurrency: 8,
		DeviceMemory:        16,
		TimeZone:            "America/New_York",
		Languages:           []string{"en-US", "en"},
		UserAgent:           "Mozilla/5.0 Chrome/120.0.0.0",
		Platform:            "Win32",
	}

	updated := base
	updated.UserAgent = "Mozilla/5.0 Chrome/121.0.0.0" // Browser update

	v1 := calc.ExtractFeatures(base)
	v2 := calc.ExtractFeatures(updated)

	similarity := calc.JaccardSimilarity(v1, v2)

	// Should still be high (>0.75) since only low-weight software changed
	if similarity < 0.75 {
		t.Errorf("Expected similarity ≥0.75 after browser update, got %.2f", similarity)
	}
}

func TestJaccardSimilarity_HardwareChange(t *testing.T) {
	calc := NewCalculator(DefaultWeights)

	base := models.Signals{
		Canvas2DHash:        "abc123",
		AudioHash:           "def456",
		WebGLVendor:         "NVIDIA",
		WebGLRenderer:       "GeForce GTX 1080",
		HardwareConcurrency: 8,
		TimeZone:            "America/New_York",
	}

	different := base
	different.Canvas2DHash = "xyz789"
	different.AudioHash = "uvw012"
	different.WebGLRenderer = "GeForce RTX 3080" // Different GPU

	v1 := calc.ExtractFeatures(base)
	v2 := calc.ExtractFeatures(different)

	similarity := calc.JaccardSimilarity(v1, v2)

	// Should be low since high-weight hardware changed
	if similarity >= 0.75 {
		t.Errorf("Expected similarity <0.75 with hardware change, got %.2f", similarity)
	}
}

func TestComputeHardwareHash_Consistency(t *testing.T) {
	signals := models.Signals{
		Canvas2DHash:        "abc123",
		AudioHash:           "def456",
		WebGLVendor:         "NVIDIA",
		WebGLRenderer:       "GeForce GTX 1080",
		HardwareConcurrency: 8,
		DeviceMemory:        16,
	}

	hash1 := ComputeHardwareHash(signals)
	hash2 := ComputeHardwareHash(signals)

	if hash1 != hash2 {
		t.Errorf("Hardware hash should be consistent: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Error("Hardware hash should not be empty")
	}
}

func TestExtractBrowserVersion(t *testing.T) {
	tests := []struct {
		ua       string
		expected string
	}{
		{
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
			expected: "chrome:120",
		},
		{
			ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Firefox/121.0",
			expected: "firefox:121",
		},
		{
			ua:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		result := extractBrowserVersion(tt.ua)
		if result != tt.expected {
			t.Errorf("extractBrowserVersion(%q) = %q, want %q", tt.ua, result, tt.expected)
		}
	}
}
