package services

import (
	"context"
	"testing"
	"time"

	"grpc-firstls/internal/config"
	"grpc-firstls/internal/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of redis.ClientInterface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	args := m.Called(ctx, key, window)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) GetRateLimitCount(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func createTestRateLimitService() (*RateLimitService, *MockRedisClient) {
	mockRedisClient := &MockRedisClient{}
	config := config.RateLimitConfig{
		DefaultRequests: 100,
		DefaultWindow:   time.Hour,
	}
	service := NewRateLimitService(mockRedisClient, config)
	return service, mockRedisClient
}

func createTestAPIKeyForRateLimitService() *database.APIKey {
	return &database.APIKey{
		ID:                     "test-id-123",
		KeyHash:                "test-hash-abc123",
		Name:                   "Test API Key",
		RateLimitRequests:      10,
		RateLimitWindowSeconds: 60,
		IsActive:               true,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}

func createTestAPIKeyWithDefaultsForRateLimit() *database.APIKey {
	return &database.APIKey{
		ID:                     "test-id-456",
		KeyHash:                "test-hash-def456",
		Name:                   "Test API Key with Defaults",
		RateLimitRequests:      0, // Will use default
		RateLimitWindowSeconds: 0, // Will use default
		IsActive:               true,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}

func TestRateLimitService_CheckRateLimit_Allowed(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is 5, limit is 10
	mockRedisClient.On("IncrementRateLimit", ctx, "rate_limit:test-id-123", time.Duration(60)*time.Second).Return(int64(5), nil)

	// Call the method
	result, err := service.CheckRateLimit(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(5), result.Remaining) // 10 - 5 = 5
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_CheckRateLimit_Exceeded(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is 11, limit is 10
	mockRedisClient.On("IncrementRateLimit", ctx, "rate_limit:test-id-123", time.Duration(60)*time.Second).Return(int64(11), nil)

	// Call the method
	result, err := service.CheckRateLimit(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(0), result.Remaining) // Should be 0 when exceeded
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_CheckRateLimit_WithDefaults(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data with default values
	testAPIKey := createTestAPIKeyWithDefaultsForRateLimit()
	ctx := context.Background()

	// Setup mock expectations - should use default window (1 hour)
	mockRedisClient.On("IncrementRateLimit", ctx, "rate_limit:test-id-456", time.Hour).Return(int64(50), nil)

	// Call the method
	result, err := service.CheckRateLimit(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(100), result.Limit)    // Should use default limit
	assert.Equal(t, int64(50), result.Remaining) // 100 - 50 = 50

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_CheckRateLimit_RedisError(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock to return error
	mockRedisClient.On("IncrementRateLimit", ctx, "rate_limit:test-id-123", time.Duration(60)*time.Second).Return(int64(0), assert.AnError)

	// Call the method
	result, err := service.CheckRateLimit(ctx, testAPIKey)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check rate limit")

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_GetRateLimitStatus_Allowed(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is 3, limit is 10
	mockRedisClient.On("GetRateLimitCount", ctx, "rate_limit:test-id-123").Return(int64(3), nil)

	// Call the method
	result, err := service.GetRateLimitStatus(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(7), result.Remaining) // 10 - 3 = 7
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_GetRateLimitStatus_Exceeded(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is 12, limit is 10
	mockRedisClient.On("GetRateLimitCount", ctx, "rate_limit:test-id-123").Return(int64(12), nil)

	// Call the method
	result, err := service.GetRateLimitStatus(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(0), result.Remaining) // Should be 0 when exceeded
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_GetRateLimitStatus_KeyNotFound(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - key doesn't exist (error returned)
	mockRedisClient.On("GetRateLimitCount", ctx, "rate_limit:test-id-123").Return(int64(0), assert.AnError)

	// Call the method
	result, err := service.GetRateLimitStatus(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err) // Should not return error, just treat as 0 count
	assert.NotNil(t, result)
	assert.True(t, result.Allowed) // Should be allowed with 0 count
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(10), result.Remaining) // 10 - 0 = 10
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_GetRateLimitStatus_WithDefaults(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data with default values
	testAPIKey := createTestAPIKeyWithDefaultsForRateLimit()
	ctx := context.Background()

	// Setup mock expectations
	mockRedisClient.On("GetRateLimitCount", ctx, "rate_limit:test-id-456").Return(int64(25), nil)

	// Call the method
	result, err := service.GetRateLimitStatus(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(100), result.Limit)    // Should use default limit
	assert.Equal(t, int64(75), result.Remaining) // 100 - 25 = 75

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_CheckRateLimit_EdgeCase_ExactlyAtLimit(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is exactly at limit (10)
	mockRedisClient.On("IncrementRateLimit", ctx, "rate_limit:test-id-123", time.Duration(60)*time.Second).Return(int64(10), nil)

	// Call the method
	result, err := service.CheckRateLimit(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Allowed) // Should still be allowed at exactly the limit
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(0), result.Remaining) // 10 - 10 = 0
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_CheckRateLimit_EdgeCase_OneOverLimit(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is 1 over limit (11)
	mockRedisClient.On("IncrementRateLimit", ctx, "rate_limit:test-id-123", time.Duration(60)*time.Second).Return(int64(11), nil)

	// Call the method
	result, err := service.CheckRateLimit(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Allowed) // Should not be allowed over the limit
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(0), result.Remaining) // Should be 0 when exceeded
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}

func TestRateLimitService_GetRateLimitStatus_EdgeCase_ExactlyAtLimit(t *testing.T) {
	service, mockRedisClient := createTestRateLimitService()

	// Create test data
	testAPIKey := createTestAPIKeyForRateLimitService()
	ctx := context.Background()

	// Setup mock expectations - current count is exactly at limit (10)
	mockRedisClient.On("GetRateLimitCount", ctx, "rate_limit:test-id-123").Return(int64(10), nil)

	// Call the method
	result, err := service.GetRateLimitStatus(ctx, testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Allowed) // Should not be allowed at exactly the limit for status check
	assert.Equal(t, int64(10), result.Limit)
	assert.Equal(t, int64(0), result.Remaining) // 10 - 10 = 0
	assert.True(t, result.ResetTime.After(time.Now()))

	mockRedisClient.AssertExpectations(t)
}
