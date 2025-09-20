package services

import (
	"context"
	"fmt"
	"time"

	"grpc-firstls/internal/config"
	"grpc-firstls/internal/database"
	"grpc-firstls/internal/redis"
)

type RateLimitService struct {
	redisClient *redis.Client
	config      config.RateLimitConfig
}

func NewRateLimitService(redisClient *redis.Client, config config.RateLimitConfig) *RateLimitService {
	return &RateLimitService{
		redisClient: redisClient,
		config:      config,
	}
}

type RateLimitResult struct {
	Allowed      bool
	Remaining    int64
	ResetTime    time.Time
	Limit        int64
}

func (s *RateLimitService) CheckRateLimit(ctx context.Context, apiKey *database.APIKey) (*RateLimitResult, error) {
	// Use API key ID as the Redis key
	redisKey := fmt.Sprintf("rate_limit:%s", apiKey.ID)
	
	// Get rate limit configuration from API key or use defaults
	limit := int64(apiKey.RateLimitRequests)
	window := time.Duration(apiKey.RateLimitWindowSeconds) * time.Second
	
	// If API key doesn't have specific limits, use defaults
	if limit <= 0 {
		limit = int64(s.config.DefaultRequests)
	}
	if window <= 0 {
		window = s.config.DefaultWindow
	}
	
	// Increment counter and get current count
	currentCount, err := s.redisClient.IncrementRateLimit(ctx, redisKey, window)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}
	
	// Check if limit exceeded
	allowed := currentCount <= limit
	remaining := limit - currentCount
	if remaining < 0 {
		remaining = 0
	}
	
	// Calculate reset time
	resetTime := time.Now().Add(window)
	
	return &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
		Limit:     limit,
	}, nil
}

func (s *RateLimitService) GetRateLimitStatus(ctx context.Context, apiKey *database.APIKey) (*RateLimitResult, error) {
	redisKey := fmt.Sprintf("rate_limit:%s", apiKey.ID)
	
	// Get current count without incrementing
	currentCount, err := s.redisClient.GetRateLimitCount(ctx, redisKey)
	if err != nil {
		// If key doesn't exist, count is 0
		currentCount = 0
	}
	
	// Get rate limit configuration
	limit := int64(apiKey.RateLimitRequests)
	window := time.Duration(apiKey.RateLimitWindowSeconds) * time.Second
	
	if limit <= 0 {
		limit = int64(s.config.DefaultRequests)
	}
	if window <= 0 {
		window = s.config.DefaultWindow
	}
	
	allowed := currentCount < limit
	remaining := limit - currentCount
	if remaining < 0 {
		remaining = 0
	}
	
	resetTime := time.Now().Add(window)
	
	return &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
		Limit:     limit,
	}, nil
}
