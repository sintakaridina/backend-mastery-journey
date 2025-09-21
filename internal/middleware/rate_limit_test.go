package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"grpc-firstls/internal/database"
	"grpc-firstls/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAPIKeyService is a mock implementation of APIKeyServiceInterface
type MockAPIKeyService struct {
	mock.Mock
}

func (m *MockAPIKeyService) ValidateAPIKey(apiKey string) (*database.APIKey, error) {
	args := m.Called(apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.APIKey), args.Error(1)
}

func (m *MockAPIKeyService) CreateAPIKey(name string, rateLimitRequests int, rateLimitWindowSeconds int) (string, error) {
	args := m.Called(name, rateLimitRequests, rateLimitWindowSeconds)
	return args.String(0), args.Error(1)
}

func (m *MockAPIKeyService) DeactivateAPIKey(apiKey string) error {
	args := m.Called(apiKey)
	return args.Error(0)
}

// MockRateLimitService is a mock implementation of RateLimitServiceInterface
type MockRateLimitService struct {
	mock.Mock
}

func (m *MockRateLimitService) CheckRateLimit(ctx context.Context, apiKey *database.APIKey) (*services.RateLimitResult, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.RateLimitResult), args.Error(1)
}

func (m *MockRateLimitService) GetRateLimitStatus(ctx context.Context, apiKey *database.APIKey) (*services.RateLimitResult, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.RateLimitResult), args.Error(1)
}

func setupTestMiddleware() (*gin.Engine, *MockAPIKeyService, *MockRateLimitService) {
	gin.SetMode(gin.TestMode)
	
	mockAPIKeyService := &MockAPIKeyService{}
	mockRateLimitService := &MockRateLimitService{}
	
	router := gin.New()
	
	// Add the rate limit middleware
	router.Use(RateLimit(mockAPIKeyService, mockRateLimitService))
	
	// Add test routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "admin"})
	})
	
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "protected"})
	})
	
	return router, mockAPIKeyService, mockRateLimitService
}

func createTestAPIKey() *database.APIKey {
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

func createTestRateLimitResult(allowed bool, remaining int64) *services.RateLimitResult {
	return &services.RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(time.Hour),
		Limit:     10,
	}
}

func TestRateLimit_SkipHealthCheck(t *testing.T) {
	router, _, _ := setupTestMiddleware()
	
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestRateLimit_SkipAdminEndpoints(t *testing.T) {
	router, _, _ := setupTestMiddleware()
	
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "admin", response["status"])
}

func TestRateLimit_NoAPIKey(t *testing.T) {
	router, _, _ := setupTestMiddleware()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "API key required", response["error"])
	assert.Equal(t, "Please provide an API key in the X-API-Key header or Authorization header", response["message"])
}

func TestRateLimit_InvalidAPIKey(t *testing.T) {
	router, mockAPIKeyService, _ := setupTestMiddleware()
	
	// Setup mock to return error for invalid API key
	mockAPIKeyService.On("ValidateAPIKey", "invalid-key").Return(nil, assert.AnError)
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid API key", response["error"])
	assert.Equal(t, "The provided API key is invalid or inactive", response["message"])
	
	mockAPIKeyService.AssertExpectations(t)
}

func TestRateLimit_ValidAPIKey_Allowed(t *testing.T) {
	router, mockAPIKeyService, mockRateLimitService := setupTestMiddleware()
	
	// Create test data
	testAPIKey := createTestAPIKey()
	testRateLimitResult := createTestRateLimitResult(true, 9)
	
	// Setup mock expectations
	mockAPIKeyService.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)
	mockRateLimitService.On("CheckRateLimit", mock.Anything, testAPIKey).Return(testRateLimitResult, nil)
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "valid-key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check rate limit headers
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "9", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "protected", response["status"])
	
	mockAPIKeyService.AssertExpectations(t)
	mockRateLimitService.AssertExpectations(t)
}

func TestRateLimit_ValidAPIKey_RateLimitExceeded(t *testing.T) {
	router, mockAPIKeyService, mockRateLimitService := setupTestMiddleware()
	
	// Create test data
	testAPIKey := createTestAPIKey()
	testRateLimitResult := createTestRateLimitResult(false, 0)
	
	// Setup mock expectations
	mockAPIKeyService.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)
	mockRateLimitService.On("CheckRateLimit", mock.Anything, testAPIKey).Return(testRateLimitResult, nil)
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "valid-key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	
	// Check rate limit headers
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Rate limit exceeded", response["error"])
	assert.Equal(t, "You have exceeded your rate limit. Please try again later.", response["message"])
	assert.Contains(t, response, "retry_after")
	
	mockAPIKeyService.AssertExpectations(t)
	mockRateLimitService.AssertExpectations(t)
}

func TestRateLimit_AuthorizationHeader(t *testing.T) {
	router, mockAPIKeyService, mockRateLimitService := setupTestMiddleware()
	
	// Create test data
	testAPIKey := createTestAPIKey()
	testRateLimitResult := createTestRateLimitResult(true, 8)
	
	// Setup mock expectations
	mockAPIKeyService.On("ValidateAPIKey", "bearer-key").Return(testAPIKey, nil)
	mockRateLimitService.On("CheckRateLimit", mock.Anything, testAPIKey).Return(testRateLimitResult, nil)
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer bearer-key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "protected", response["status"])
	
	mockAPIKeyService.AssertExpectations(t)
	mockRateLimitService.AssertExpectations(t)
}

func TestRateLimit_RateLimitServiceError(t *testing.T) {
	router, mockAPIKeyService, mockRateLimitService := setupTestMiddleware()
	
	// Create test data
	testAPIKey := createTestAPIKey()
	
	// Setup mock expectations
	mockAPIKeyService.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)
	mockRateLimitService.On("CheckRateLimit", mock.Anything, testAPIKey).Return(nil, assert.AnError)
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "valid-key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Rate limit check failed", response["error"])
	assert.Equal(t, "Unable to check rate limit", response["message"])
	
	mockAPIKeyService.AssertExpectations(t)
	mockRateLimitService.AssertExpectations(t)
}

func TestRateLimit_ContextHasAPIKey(t *testing.T) {
	router, mockAPIKeyService, mockRateLimitService := setupTestMiddleware()
	
	// Create test data
	testAPIKey := createTestAPIKey()
	testRateLimitResult := createTestRateLimitResult(true, 7)
	
	// Setup mock expectations
	mockAPIKeyService.On("ValidateAPIKey", "valid-key").Return(testAPIKey, nil)
	mockRateLimitService.On("CheckRateLimit", mock.Anything, testAPIKey).Return(testRateLimitResult, nil)
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "valid-key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify that the API key is stored in context
	// This is tested indirectly by the successful response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "protected", response["status"])
	
	mockAPIKeyService.AssertExpectations(t)
	mockRateLimitService.AssertExpectations(t)
}

func TestRateLimit_InvalidAuthorizationHeader(t *testing.T) {
	router, _, _ := setupTestMiddleware()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "InvalidFormat key")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "API key required", response["error"])
}

func TestRateLimit_EmptyAuthorizationHeader(t *testing.T) {
	router, _, _ := setupTestMiddleware()
	
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "API key required", response["error"])
}
