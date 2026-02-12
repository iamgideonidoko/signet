package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("DATABASE_URL", "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable")
	_ = os.Setenv("REDIS_URL", "redis://localhost:6379/0")
	_ = os.Setenv("API_PORT", "6969")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test API config
	if cfg.API.Port != "6969" {
		t.Errorf("Expected API port 6969, got %s", cfg.API.Port)
	}

	// Test Database config
	if cfg.Database.URL != "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("Expected DATABASE_URL to be set, got %s", cfg.Database.URL)
	}

	if cfg.Database.MaxConns != 25 {
		t.Errorf("Expected MaxConns 25, got %d", cfg.Database.MaxConns)
	}

	// Test Redis config
	if cfg.Redis.URL != "redis://localhost:6379/0" {
		t.Errorf("Expected REDIS_URL to be set, got %s", cfg.Redis.URL)
	}

	if cfg.Redis.CacheTTL != 48*time.Hour {
		t.Errorf("Expected CacheTTL 48h, got %v", cfg.Redis.CacheTTL)
	}
}

func TestConfigValidation(t *testing.T) {
	// Clear environment to test defaults
	os.Clearenv()

	cfg, err := Load()
	// Should not fail - default URL is provided
	if err != nil {
		t.Errorf("Load() with defaults should succeed, got error: %v", err)
	}

	// Verify defaults are set
	if cfg.API.Port != "6969" {
		t.Errorf("Expected default port 6969, got %s", cfg.API.Port)
	}

	if cfg.Database.URL == "" {
		t.Error("Expected default DATABASE_URL to be set")
	}

	if cfg.Redis.URL == "" {
		t.Error("Expected default REDIS_URL to be set")
	}
}
