package similarity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/iamgideonidoko/signet/internal/models"
)

// Weights for different signal categories.
type Weights struct {
	Hardware    float64
	Environment float64
	Software    float64
}

var DefaultWeights = Weights{
	Hardware:    0.8,
	Environment: 0.5,
	Software:    0.2,
}

// FeatureVector represents a fingerprint as weighted features.
type FeatureVector struct {
	Features map[string]float64
	Hash     string
}

// Calculator computes similarity between fingerprints.
type Calculator struct {
	weights Weights
}

func NewCalculator(weights Weights) *Calculator {
	return &Calculator{weights: weights}
}

// ExtractFeatures converts signals into a weighted feature vector.
func (c *Calculator) ExtractFeatures(signals models.Signals) FeatureVector {
	features := make(map[string]float64)

	// Hardware features (weight: 0.8)
	if signals.Canvas2DHash != "" {
		features["canvas:"+signals.Canvas2DHash] = c.weights.Hardware
	}
	if signals.AudioHash != "" {
		features["audio:"+signals.AudioHash] = c.weights.Hardware
	}
	features[fmt.Sprintf("webgl:%s:%s", signals.WebGLVendor, signals.WebGLRenderer)] = c.weights.Hardware

	// WebGL extensions (sorted for consistency)
	extHash := hashStringSlice(signals.WebGLExtensions)
	features["webgl_ext:"+extHash] = c.weights.Hardware * 0.7

	features[fmt.Sprintf("hw_concurrency:%d", signals.HardwareConcurrency)] = c.weights.Hardware * 0.6
	features[fmt.Sprintf("device_memory:%.0f", signals.DeviceMemory)] = c.weights.Hardware * 0.6
	features[fmt.Sprintf("color_depth:%d", signals.ColorDepth)] = c.weights.Hardware * 0.5

	// Environment features (weight: 0.5)
	if signals.TimeZone != "" {
		features["tz:"+signals.TimeZone] = c.weights.Environment
	}
	langHash := hashStringSlice(signals.Languages)
	features["lang:"+langHash] = c.weights.Environment

	fontHash := hashStringSlice(signals.Fonts)
	features["fonts:"+fontHash] = c.weights.Environment * 0.9

	features[fmt.Sprintf("screen:%dx%d", signals.ScreenWidth, signals.ScreenHeight)] = c.weights.Environment * 0.7

	// Software features (weight: 0.2) - most volatile
	if signals.Platform != "" {
		features["platform:"+signals.Platform] = c.weights.Software
	}

	// Extract browser/version from UA (ignore patch versions)
	browserVersion := extractBrowserVersion(signals.UserAgent)
	if browserVersion != "" {
		features["browser:"+browserVersion] = c.weights.Software
	}

	// Generate overall hash for quick lookups
	hash := c.computeVectorHash(features)

	return FeatureVector{
		Features: features,
		Hash:     hash,
	}
}

// ComputeHardwareHash generates a hash from hardware-only signals for Redis caching.
func ComputeHardwareHash(signals models.Signals) string {
	parts := []string{
		signals.Canvas2DHash,
		signals.AudioHash,
		signals.WebGLVendor,
		signals.WebGLRenderer,
		fmt.Sprintf("%d", signals.HardwareConcurrency),
		fmt.Sprintf("%.0f", signals.DeviceMemory),
	}

	combined := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// JaccardSimilarity computes weighted Jaccard similarity between two feature vectors.
func (c *Calculator) JaccardSimilarity(v1, v2 FeatureVector) float64 {
	if len(v1.Features) == 0 || len(v2.Features) == 0 {
		return 0.0
	}

	// Quick check: if hardware hashes match exactly, return high similarity
	if v1.Hash == v2.Hash {
		return 1.0
	}

	var intersection, union float64

	allKeys := make(map[string]bool)
	for k := range v1.Features {
		allKeys[k] = true
	}
	for k := range v2.Features {
		allKeys[k] = true
	}

	for key := range allKeys {
		w1, exists1 := v1.Features[key]
		w2, exists2 := v2.Features[key]

		switch {
		case exists1 && exists2:
			intersection += math.Min(w1, w2)
			union += math.Max(w1, w2)
		case exists1:
			union += w1
		default:
			union += w2
		}
	}

	if union == 0 {
		return 0.0
	}

	return intersection / union
}

// computeVectorHash creates a deterministic hash of the feature vector.
func (c *Calculator) computeVectorHash(features map[string]float64) string {
	// Sort keys for deterministic hashing
	keys := make([]string, 0, len(features))
	for k := range features {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		// Only include high-weight features in the hash
		if features[k] >= c.weights.Environment {
			parts = append(parts, fmt.Sprintf("%s:%.2f", k, features[k]))
		}
	}

	combined := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// hashStringSlice creates a consistent hash from a string slice.
func hashStringSlice(items []string) string {
	if len(items) == 0 {
		return "empty"
	}

	sorted := make([]string, len(items))
	copy(sorted, items)
	sort.Strings(sorted)

	combined := strings.Join(sorted, ",")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:8])
}

// extractBrowserVersion extracts browser name and major version from UA.
func extractBrowserVersion(ua string) string {
	if ua == "" {
		return ""
	}

	// Simple extraction - can be enhanced with a proper UA parser
	ua = strings.ToLower(ua)

	browsers := []string{"chrome", "firefox", "safari", "edge", "opera"}
	for _, browser := range browsers {
		if idx := strings.Index(ua, browser); idx != -1 {
			// Extract major version only
			parts := strings.Split(ua[idx:], "/")
			if len(parts) > 1 {
				version := strings.Split(parts[1], ".")[0]
				return browser + ":" + strings.TrimSpace(version)
			}
		}
	}

	return "unknown"
}
