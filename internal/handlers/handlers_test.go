package handlers

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

func setupTestRouter() (*gin.Engine, *MockAPIKeyService, *MockRateLimitService, *Handler) {
	gin.SetMode(gin.TestMode)

	mockAPIKeyService := &MockAPIKeyService{}
	mockRateLimitService := &MockRateLimitService{}
	handler := NewHandler(mockAPIKeyService, mockRateLimitService)

	router := gin.New()
	handler.SetupRoutes(router)

	return router, mockAPIKeyService, mockRateLimitService, handler
}

func createTestAPIKey() *database.APIKey {
	return &database.APIKey{
		ID:                     "test-id-123",
		KeyHash:                "test-hash",
		Name:                   "Test API Key",
		RateLimitRequests:      100,
		RateLimitWindowSeconds: 3600,
		IsActive:               true,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}

func createTestRateLimitResult() *services.RateLimitResult {
	return &services.RateLimitResult{
		Allowed:   true,
		Remaining: 99,
		ResetTime: time.Now().Add(time.Hour),
		Limit:     100,
	}
}

func TestHealthCheck(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "rate-limiter-api", response["service"])
}

func TestCreateAPIKey_Success(t *testing.T) {
	router, mockAPIKeyService, _, _ := setupTestRouter()

	// Setup mock expectations
	expectedAPIKey := "ak_1234567890_abcdef"
	mockAPIKeyService.On("CreateAPIKey", "Test API Key", 100, 3600).Return(expectedAPIKey, nil)

	// Create request body
	requestBody := map[string]interface{}{
		"name":                      "Test API Key",
		"rate_limit_requests":       100,
		"rate_limit_window_seconds": 3600,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedAPIKey, response["api_key"])
	assert.Equal(t, "Test API Key", response["name"])

	rateLimit := response["rate_limit"].(map[string]interface{})
	assert.Equal(t, float64(100), rateLimit["requests"])
	assert.Equal(t, float64(3600), rateLimit["window_seconds"])

	mockAPIKeyService.AssertExpectations(t)
}

func TestCreateAPIKey_WithDefaults(t *testing.T) {
	router, mockAPIKeyService, _, _ := setupTestRouter()

	// Setup mock expectations with default values
	expectedAPIKey := "ak_1234567890_abcdef"
	mockAPIKeyService.On("CreateAPIKey", "Test API Key", 100, 3600).Return(expectedAPIKey, nil)

	// Create request body without rate limit fields
	requestBody := map[string]interface{}{
		"name": "Test API Key",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, expectedAPIKey, response["api_key"])

	rateLimit := response["rate_limit"].(map[string]interface{})
	assert.Equal(t, float64(100), rateLimit["requests"])        // Default value
	assert.Equal(t, float64(3600), rateLimit["window_seconds"]) // Default value

	mockAPIKeyService.AssertExpectations(t)
}

func TestCreateAPIKey_InvalidRequest(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	// Create invalid request body (missing required name field)
	requestBody := map[string]interface{}{
		"rate_limit_requests": 100,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Invalid request", response["error"])
	assert.Contains(t, response["message"], "Name")
}

func TestCreateAPIKey_ServiceError(t *testing.T) {
	router, mockAPIKeyService, _, _ := setupTestRouter()

	// Setup mock to return error
	mockAPIKeyService.On("CreateAPIKey", "Test API Key", 100, 3600).Return("", fmt.Errorf("database error"))

	requestBody := map[string]interface{}{
		"name":                      "Test API Key",
		"rate_limit_requests":       100,
		"rate_limit_window_seconds": 3600,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Failed to create API key", response["error"])
	assert.Equal(t, "database error", response["message"])

	mockAPIKeyService.AssertExpectations(t)
}

func TestDeactivateAPIKey_Success(t *testing.T) {
	router, mockAPIKeyService, _, _ := setupTestRouter()

	// Setup mock expectations
	testAPIKey := "ak_1234567890_abcdef"
	mockAPIKeyService.On("DeactivateAPIKey", testAPIKey).Return(nil)

	req, _ := http.NewRequest("DELETE", "/admin/api-keys/"+testAPIKey, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "API key deactivated successfully", response["message"])

	mockAPIKeyService.AssertExpectations(t)
}

func TestDeactivateAPIKey_MissingKey(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	// Test with empty key in URL path
	req, _ := http.NewRequest("DELETE", "/admin/api-keys/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// This should return 404 because the route doesn't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeactivateAPIKey_NotFound(t *testing.T) {
	router, mockAPIKeyService, _, _ := setupTestRouter()

	// Setup mock to return error
	testAPIKey := "ak_1234567890_abcdef"
	mockAPIKeyService.On("DeactivateAPIKey", testAPIKey).Return(fmt.Errorf("API key not found"))

	req, _ := http.NewRequest("DELETE", "/admin/api-keys/"+testAPIKey, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "API key not found", response["error"])
	assert.Equal(t, "API key not found", response["message"])

	mockAPIKeyService.AssertExpectations(t)
}

func TestGetStatus_Success(t *testing.T) {
	// Create a test API key
	testAPIKey := createTestAPIKey()

	// Create request with API key in context (simulating middleware)
	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	// Create a new context with API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("api_key", testAPIKey)

	// Create handler and call directly
	_, _, _, handler := setupTestRouter()
	handler.GetStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "authenticated", response["status"])

	apiKeyInfo := response["api_key"].(map[string]interface{})
	assert.Equal(t, "test-id-123", apiKeyInfo["id"])
	assert.Equal(t, "Test API Key", apiKeyInfo["name"])
}

func TestGetStatus_NoAPIKeyInContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	// Create context without API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create handler and call directly
	_, _, _, handler := setupTestRouter()
	handler.GetStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "API key not found in context", response["error"])
}

func TestGetRateLimitStatus_Success(t *testing.T) {
	// Create test data
	testAPIKey := createTestAPIKey()
	testRateLimitResult := createTestRateLimitResult()

	// Setup mock expectations
	_, _, mockRateLimitService, handler := setupTestRouter()
	mockRateLimitService.On("GetRateLimitStatus", mock.Anything, testAPIKey).Return(testRateLimitResult, nil)

	req, _ := http.NewRequest("GET", "/api/rate-limit", nil)
	w := httptest.NewRecorder()

	// Create context with API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("api_key", testAPIKey)

	// Call handler directly
	handler.GetRateLimitStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	rateLimit := response["rate_limit"].(map[string]interface{})
	assert.Equal(t, float64(100), rateLimit["limit"])
	assert.Equal(t, float64(99), rateLimit["remaining"])
	assert.Equal(t, true, rateLimit["allowed"])

	mockRateLimitService.AssertExpectations(t)
}

func TestGetRateLimitStatus_ServiceError(t *testing.T) {
	// Create test data
	testAPIKey := createTestAPIKey()

	// Setup mock to return error
	_, _, mockRateLimitService, handler := setupTestRouter()
	mockRateLimitService.On("GetRateLimitStatus", mock.Anything, testAPIKey).Return(nil, fmt.Errorf("redis error"))

	req, _ := http.NewRequest("GET", "/api/rate-limit", nil)
	w := httptest.NewRecorder()

	// Create context with API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("api_key", testAPIKey)

	// Call handler directly
	handler.GetRateLimitStatus(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Failed to get rate limit status", response["error"])
	assert.Equal(t, "redis error", response["message"])

	mockRateLimitService.AssertExpectations(t)
}

func TestTestEndpoint_Success(t *testing.T) {
	// Create test data
	testAPIKey := createTestAPIKey()

	// Create request body
	requestBody := map[string]interface{}{
		"message": "Hello, World!",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create context with API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("api_key", testAPIKey)

	// Create handler and call directly
	_, _, _, handler := setupTestRouter()
	handler.TestEndpoint(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Request processed successfully", response["message"])
	assert.Equal(t, "Hello, World!", response["echo"])

	apiKeyInfo := response["api_key"].(map[string]interface{})
	assert.Equal(t, "test-id-123", apiKeyInfo["id"])
	assert.Equal(t, "Test API Key", apiKeyInfo["name"])
}

func TestTestEndpoint_InvalidJSON(t *testing.T) {
	// Create test data
	testAPIKey := createTestAPIKey()

	// Create invalid JSON
	req, _ := http.NewRequest("POST", "/api/test", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create context with API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("api_key", testAPIKey)

	// Create handler and call directly
	_, _, _, handler := setupTestRouter()
	handler.TestEndpoint(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Invalid request", response["error"])
}

func TestTestEndpoint_NoAPIKeyInContext(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/test", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create context without API key
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create handler and call directly
	_, _, _, handler := setupTestRouter()
	handler.TestEndpoint(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "API key not found in context", response["error"])
}
