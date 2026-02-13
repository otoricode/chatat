// Package config handles application configuration loading from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds all configuration values for the application.
type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Environment string // development, staging, production

	// WhatsApp (GOWA) configuration
	WABaseURL       string // e.g. http://localhost:3000
	WAWebhookSecret string // HMAC secret for webhook signature verification
	WABusinessPhone string // WhatsApp business phone number (E.164)

	// S3-compatible storage configuration
	S3Endpoint  string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3Region    string
}

// Load reads configuration from environment variables and returns a Config.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		Environment: getEnv("ENVIRONMENT", "development"),

		WABaseURL:       getEnv("WA_BASE_URL", "http://localhost:3000"),
		WAWebhookSecret: getEnv("WA_WEBHOOK_SECRET", "chatat-webhook-secret"),
		WABusinessPhone: getEnv("WA_BUSINESS_PHONE", ""),

		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3Bucket:    getEnv("S3_BUCKET", "chatat-media"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
