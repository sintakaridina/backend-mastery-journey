package services

import (
	"context"

	"grpc-firstls/internal/database"
)

// APIKeyServiceInterface defines the interface for API key operations
type APIKeyServiceInterface interface {
	ValidateAPIKey(apiKey string) (*database.APIKey, error)
	CreateAPIKey(name string, rateLimitRequests int, rateLimitWindowSeconds int) (string, error)
	DeactivateAPIKey(apiKey string) error
}

// RateLimitServiceInterface defines the interface for rate limiting operations
type RateLimitServiceInterface interface {
	CheckRateLimit(ctx context.Context, apiKey *database.APIKey) (*RateLimitResult, error)
	GetRateLimitStatus(ctx context.Context, apiKey *database.APIKey) (*RateLimitResult, error)
}
