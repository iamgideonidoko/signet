package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	API         APIConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Fingerprint FingerprintConfig
	RateLimit   RateLimitConfig
	Security    SecurityConfig
	Monitoring  MonitoringConfig
}

type APIConfig struct {
	Port        string
	Host        string
	Environment string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxConns     int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	CacheTTL time.Duration
}

type FingerprintConfig struct {
	SimilarityThreshold float64
	HardwareWeight      float64
	EnvironmentWeight   float64
	SoftwareWeight      float64
}

type RateLimitConfig struct {
	Requests           int
	Window             time.Duration
	RequestsByHardware int
	HardwareWindow     time.Duration
}

type SecurityConfig struct {
	CORSOrigins    []string
	TrustedProxies []string
}

type MonitoringConfig struct {
	EnableMetrics bool
	LogLevel      string
}

func Load() (*Config, error) {
	cfg := &Config{
		API: APIConfig{
			Port:        getEnv("API_PORT", "8080"),
			Host:        getEnv("API_HOST", "0.0.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "signet"),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "signet"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxConns:     getEnvInt("DB_MAX_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			CacheTTL: getEnvDuration("REDIS_CACHE_TTL", 48*time.Hour),
		},
		Fingerprint: FingerprintConfig{
			SimilarityThreshold: getEnvFloat("SIMILARITY_THRESHOLD", 0.75),
			HardwareWeight:      getEnvFloat("HARDWARE_WEIGHT", 0.8),
			EnvironmentWeight:   getEnvFloat("ENVIRONMENT_WEIGHT", 0.5),
			SoftwareWeight:      getEnvFloat("SOFTWARE_WEIGHT", 0.2),
		},
		RateLimit: RateLimitConfig{
			Requests:           getEnvInt("RATE_LIMIT_REQUESTS", 1000),
			Window:             getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
			RequestsByHardware: getEnvInt("RATE_LIMIT_BY_HARDWARE", 2000),
			HardwareWindow:     getEnvDuration("RATE_LIMIT_HARDWARE_WINDOW", 1*time.Hour),
		},
		Security: SecurityConfig{
			CORSOrigins:    getEnvSlice("CORS_ORIGINS", []string{"*"}),
			TrustedProxies: getEnvSlice("TRUSTED_PROXIES", []string{}),
		},
		Monitoring: MonitoringConfig{
			EnableMetrics: getEnvBool("ENABLE_METRICS", true),
			LogLevel:      getEnv("LOG_LEVEL", "info"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.Fingerprint.SimilarityThreshold < 0 || c.Fingerprint.SimilarityThreshold > 1 {
		return fmt.Errorf("SIMILARITY_THRESHOLD must be between 0 and 1")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		var result []string
		for _, item := range splitAndTrim(value, ",") {
			if item != "" {
				result = append(result, item)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range splitString(s, sep) {
		result = append(result, trimSpace(item))
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := []string{}
	current := ""
	for _, r := range s {
		if string(r) == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	parts = append(parts, current)
	return parts
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
