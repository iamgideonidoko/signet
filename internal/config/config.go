package config

import "os"

type APIConfig struct {
	Port        string
	Host        string
	Environment string
}

type Config struct {
	API APIConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		API: APIConfig{
			Port:        getEnv("API_PORT", "6969"),
			Host:        getEnv("API_HOST", "0.0.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	return nil
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
