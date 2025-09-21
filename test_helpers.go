package main

import (
	"context"
	"database/sql"
	"time"

	"grpc-firstls/internal/database"
	"grpc-firstls/internal/services"
)

// TestHelper provides utility functions for testing
type TestHelper struct{}

// CreateTestAPIKey creates a test API key for testing purposes
func (th *TestHelper) CreateTestAPIKey() *database.APIKey {
	return &database.APIKey{
		ID:                      "test-id-123",
		KeyHash:                 "test-hash-abc123",
		Name:                    "Test API Key",
		RateLimitRequests:       10,
		RateLimitWindowSeconds:  60,
		IsActive:                true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
}

// CreateTestRateLimitResult creates a test rate limit result
func (th *TestHelper) CreateTestRateLimitResult(allowed bool, remaining int64) *services.RateLimitResult {
	return &services.RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(time.Hour),
		Limit:     10,
	}
}

// MockDB is a mock implementation of database.DB for testing
type MockDB struct {
	apiKeys map[string]*database.APIKey
}

func NewMockDB() *MockDB {
	return &MockDB{
		apiKeys: make(map[string]*database.APIKey),
	}
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	// Mock implementation - in real tests, you'd use a proper mock
	return nil
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Mock implementation - in real tests, you'd use a proper mock
	return nil, nil
}

// MockRedisClient is a mock implementation of redis.Client for testing
type MockRedisClient struct {
	counters map[string]int64
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		counters: make(map[string]int64),
	}
}

func (m *MockRedisClient) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	m.counters[key]++
	return m.counters[key], nil
}

func (m *MockRedisClient) GetRateLimitCount(ctx context.Context, key string) (int64, error) {
	return m.counters[key], nil
}

// TestData provides test data for various scenarios
type TestData struct{}

func (td *TestData) GetValidAPIKeyRequest() map[string]interface{} {
	return map[string]interface{}{
		"name":                      "Test API Key",
		"rate_limit_requests":       100,
		"rate_limit_window_seconds": 3600,
	}
}

func (td *TestData) GetInvalidAPIKeyRequest() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit_requests": 100,
		// Missing required "name" field
	}
}

func (td *TestData) GetTestEndpointRequest() map[string]interface{} {
	return map[string]interface{}{
		"message": "Test message",
	}
}

func (td *TestData) GetLowRateLimitAPIKeyRequest() map[string]interface{} {
	return map[string]interface{}{
		"name":                      "Low Rate Limit Key",
		"rate_limit_requests":       2,
		"rate_limit_window_seconds": 60,
	}
}
