package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"grpc-firstls/internal/database"
	"grpc-firstls/internal/handlers"
	"grpc-firstls/internal/middleware"
	"grpc-firstls/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSetup contains all the components needed for integration testing
type IntegrationTestSetup struct {
	Router           *gin.Engine
	APIKeyService    services.APIKeyServiceInterface
	RateLimitService services.RateLimitServiceInterface
	Handler          *handlers.Handler
	DB               *MockDB
	RedisClient      *MockRedisClient
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSetup {
	gin.SetMode(gin.TestMode)

	// For integration tests, we would normally use test databases
	// For this example, we'll use mocks but structure it like real integration
	mockDB := NewMockDB()
	mockRedisClient := NewMockRedisClient()

	// Create mock services with shared state
	apiKeyService := NewMockAPIKeyService()
	rateLimitService := NewMockRateLimitService()

	// Create handler
	handler := handlers.NewHandler(apiKeyService, rateLimitService)

	// Setup router
	router := gin.New()
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(apiKeyService, rateLimitService))
	handler.SetupRoutes(router)

	return &IntegrationTestSetup{
		Router:           router,
		APIKeyService:    apiKeyService,
		RateLimitService: rateLimitService,
		Handler:          handler,
		DB:               mockDB,
		RedisClient:      mockRedisClient,
	}
}

// Mock implementations for integration testing
// Note: MockDB and MockRedisClient are defined in test_helpers.go

// MockAPIKeyService for integration testing
type MockAPIKeyService struct {
	apiKeys map[string]*database.APIKey
}

func NewMockAPIKeyService() *MockAPIKeyService {
	return &MockAPIKeyService{
		apiKeys: make(map[string]*database.APIKey),
	}
}

func (m *MockAPIKeyService) ValidateAPIKey(apiKey string) (*database.APIKey, error) {
	// Check if the API key exists in our mock storage
	if storedKey, exists := m.apiKeys[apiKey]; exists {
		if !storedKey.IsActive {
			return nil, fmt.Errorf("API key is inactive")
		}
		return storedKey, nil
	}

	// Fallback for any key that starts with "ak_" (for backward compatibility)
	if len(apiKey) > 3 && apiKey[:3] == "ak_" {
		return &database.APIKey{
			ID:                     "test-id-123",
			KeyHash:                "test-hash",
			Name:                   "Test API Key",
			RateLimitRequests:      5,
			RateLimitWindowSeconds: 60,
			IsActive:               true,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("invalid API key")
}

func (m *MockAPIKeyService) CreateAPIKey(name string, rateLimitRequests int, rateLimitWindowSeconds int) (string, error) {
	// Generate a mock API key
	apiKey := fmt.Sprintf("ak_%d_%x", time.Now().Unix(), time.Now().UnixNano())

	// Store the API key in our mock storage
	m.apiKeys[apiKey] = &database.APIKey{
		ID:                     fmt.Sprintf("id_%d", time.Now().UnixNano()),
		KeyHash:                "mock-hash",
		Name:                   name,
		RateLimitRequests:      rateLimitRequests,
		RateLimitWindowSeconds: rateLimitWindowSeconds,
		IsActive:               true,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	return apiKey, nil
}

func (m *MockAPIKeyService) DeactivateAPIKey(apiKey string) error {
	// Check if the API key exists in our mock storage
	if storedKey, exists := m.apiKeys[apiKey]; exists {
		storedKey.IsActive = false
		return nil
	}

	// For backward compatibility, always succeed
	return nil
}

// MockRateLimitService for integration testing
type MockRateLimitService struct {
	counters map[string]int64
}

func NewMockRateLimitService() *MockRateLimitService {
	return &MockRateLimitService{
		counters: make(map[string]int64),
	}
}

func (m *MockRateLimitService) CheckRateLimit(ctx context.Context, apiKey *database.APIKey) (*services.RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:%s", apiKey.ID)
	m.counters[key]++

	limit := int64(apiKey.RateLimitRequests)
	allowed := m.counters[key] <= limit
	remaining := limit - m.counters[key]
	if remaining < 0 {
		remaining = 0
	}

	return &services.RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(time.Duration(apiKey.RateLimitWindowSeconds) * time.Second),
		Limit:     limit,
	}, nil
}

func (m *MockRateLimitService) GetRateLimitStatus(ctx context.Context, apiKey *database.APIKey) (*services.RateLimitResult, error) {
	key := fmt.Sprintf("rate_limit:%s", apiKey.ID)
	currentCount := m.counters[key]

	limit := int64(apiKey.RateLimitRequests)
	allowed := currentCount < limit
	remaining := limit - currentCount
	if remaining < 0 {
		remaining = 0
	}

	return &services.RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(time.Duration(apiKey.RateLimitWindowSeconds) * time.Second),
		Limit:     limit,
	}, nil
}

func TestIntegration_CreateAPIKeyAndUseIt(t *testing.T) {
	setup := setupIntegrationTest(t)

	// Step 1: Create an API key
	createRequest := map[string]interface{}{
		"name":                      "Integration Test Key",
		"rate_limit_requests":       5,
		"rate_limit_window_seconds": 60,
	}

	jsonBody, _ := json.Marshal(createRequest)
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	setup.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	apiKey := createResponse["api_key"].(string)
	require.NotEmpty(t, apiKey)

	// Step 2: Use the API key to access protected endpoint
	req, _ = http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()

	setup.Router.ServeHTTP(w, req)

	// This should work if the middleware is properly integrated
	// Note: In a real integration test, you'd need to mock the database properly
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_RateLimitEnforcement(t *testing.T) {
	setup := setupIntegrationTest(t)

	// Create an API key with low rate limit
	createRequest := map[string]interface{}{
		"name":                      "Rate Limit Test Key",
		"rate_limit_requests":       2,
		"rate_limit_window_seconds": 60,
	}

	jsonBody, _ := json.Marshal(createRequest)
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	setup.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	apiKey := createResponse["api_key"].(string)

	// Make requests up to the limit
	for i := 0; i < 2; i++ {
		req, _ = http.NewRequest("GET", "/api/status", nil)
		req.Header.Set("X-API-Key", apiKey)
		w = httptest.NewRecorder()

		setup.Router.ServeHTTP(w, req)

		// First two requests should succeed
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Third request should be rate limited
	req, _ = http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()

	setup.Router.ServeHTTP(w, req)

	// This should be rate limited
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var rateLimitResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &rateLimitResponse)
	require.NoError(t, err)

	assert.Equal(t, "Rate limit exceeded", rateLimitResponse["error"])
}

func TestIntegration_AdminEndpointsNotRateLimited(t *testing.T) {
	setup := setupIntegrationTest(t)

	// Test that admin endpoints don't require API keys
	req, _ := http.NewRequest("GET", "/admin/api-keys", nil)
	w := httptest.NewRecorder()

	setup.Router.ServeHTTP(w, req)

	// Should not be rate limited (though it might return 404 or method not allowed)
	// The important thing is it's not returning 401 Unauthorized
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_HealthCheckNotRateLimited(t *testing.T) {
	setup := setupIntegrationTest(t)

	// Test that health check doesn't require API keys
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
}

func TestIntegration_CompleteWorkflow(t *testing.T) {
	setup := setupIntegrationTest(t)

	// Step 1: Check health
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Step 2: Create API key
	createRequest := map[string]interface{}{
		"name":                      "Workflow Test Key",
		"rate_limit_requests":       3,
		"rate_limit_window_seconds": 60,
	}

	jsonBody, _ := json.Marshal(createRequest)
	req, _ = http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	apiKey := createResponse["api_key"].(string)

	// Step 3: Use API key to get status
	req, _ = http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 4: Check rate limit status
	req, _ = http.NewRequest("GET", "/api/rate-limit", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 5: Use test endpoint
	testRequest := map[string]interface{}{
		"message": "Integration test message",
	}

	jsonBody, _ = json.Marshal(testRequest)
	req, _ = http.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var testResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &testResponse)
	require.NoError(t, err)

	assert.Equal(t, "Request processed successfully", testResponse["message"])
	assert.Equal(t, "Integration test message", testResponse["echo"])

	// Step 6: Deactivate API key
	req, _ = http.NewRequest("DELETE", "/admin/api-keys/"+apiKey, nil)
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 7: Try to use deactivated API key
	req, _ = http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	// Should fail with invalid API key
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	setup := setupIntegrationTest(t)

	// Test invalid JSON in create API key
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test missing required fields
	createRequest := map[string]interface{}{
		"rate_limit_requests": 100,
		// Missing "name" field
	}

	jsonBody, _ := json.Marshal(createRequest)
	req, _ = http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test invalid API key format
	req, _ = http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("X-API-Key", "invalid-format")
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test missing API key
	req, _ = http.NewRequest("GET", "/api/status", nil)
	w = httptest.NewRecorder()
	setup.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
