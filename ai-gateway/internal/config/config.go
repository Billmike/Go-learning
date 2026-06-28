package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all gateway settings loaded from the environment.
type Config struct {
	Port           string
	ModelAURL      string
	ModelBURL      string
	APIToken       string
	RequestTimeout time.Duration
	RateLimit      int
	LogLevel       string
}

// Load reads configuration from environment variables, applies defaults,
// and validates required fields. It returns an error on invalid or missing
// values so the process can fail fast at startup.
func Load() (*Config, error) {
	cfg := &Config{
		Port:      envOrDefault("PORT", "8080"),
		ModelAURL: envOrDefault("MODEL_A_URL", "http://localhost:9001"),
		ModelBURL: envOrDefault("MODEL_B_URL", "http://localhost:9002"),
		LogLevel:  envOrDefault("LOG_LEVEL", "info"),
	}

	token := os.Getenv("API_TOKEN")
	if token == "" {
		return nil, errors.New("API_TOKEN is required")
	}
	cfg.APIToken = token

	timeoutStr := envOrDefault("REQUEST_TIMEOUT", "30s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REQUEST_TIMEOUT %q: %w", timeoutStr, err)
	}
	cfg.RequestTimeout = timeout

	rateLimitStr := envOrDefault("RATE_LIMIT", "100")
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT %q: %w", rateLimitStr, err)
	}
	if rateLimit <= 0 {
		return nil, errors.New("RATE_LIMIT must be a positive integer")
	}
	cfg.RateLimit = rateLimit

	return cfg, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
