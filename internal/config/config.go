package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL     string
	RedisURL        string
	RateLimitConfig RateLimitConfig
}

type RateLimitConfig struct {
	DefaultRequests int
	DefaultWindow   time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/rate_limiter?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		RateLimitConfig: RateLimitConfig{
			DefaultRequests: getEnvAsInt("DEFAULT_RATE_LIMIT_REQUESTS", 100),
			DefaultWindow:   getEnvAsDuration("DEFAULT_RATE_LIMIT_WINDOW", "1h"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}
