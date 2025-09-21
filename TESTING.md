# Testing Guide for Rate Limiter API

This document provides comprehensive information about testing the Rate Limiter API, including unit tests, integration tests, and how to run them.

## Table of Contents

- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Categories](#test-categories)
- [Test Coverage](#test-coverage)
- [Writing New Tests](#writing-new-tests)
- [Mocking](#mocking)
- [Integration Testing](#integration-testing)

## Test Structure

The test suite is organized as follows:

```
├── internal/
│   ├── handlers/
│   │   ├── handlers.go
│   │   └── handlers_test.go          # Handler unit tests
│   ├── services/
│   │   ├── api_key_service.go
│   │   ├── api_key_service_test.go   # API key service tests
│   │   ├── rate_limit_service.go
│   │   └── rate_limit_service_test.go # Rate limit service tests
│   └── middleware/
│       ├── rate_limit.go
│       ├── rate_limit_test.go        # Rate limit middleware tests
│       ├── cors.go
│       └── cors_test.go              # CORS middleware tests
├── integration_test.go               # Integration tests
├── test_helpers.go                   # Test utilities and mocks
└── Makefile                          # Test automation
```

## Running Tests

### Prerequisites

1. Install Go 1.19 or later
2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Quick Start

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
make test-coverage

# Run tests with verbose output
make test-verbose
```

### Manual Commands

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/handlers/...
go test ./internal/services/...
go test ./internal/middleware/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Test Categories

### 1. Unit Tests

Unit tests focus on testing individual components in isolation using mocks.

#### Handler Tests (`internal/handlers/handlers_test.go`)

Tests all HTTP endpoints:
- `GET /health` - Health check endpoint
- `POST /admin/api-keys` - Create API key
- `DELETE /admin/api-keys/:key` - Deactivate API key
- `GET /api/status` - Get authentication status
- `GET /api/rate-limit` - Get rate limit status
- `POST /api/test` - Test endpoint

**Key test scenarios:**
- Success cases with valid input
- Error cases with invalid input
- Missing required fields
- Service errors
- Authentication failures

#### Service Tests

**API Key Service (`internal/services/api_key_service_test.go`):**
- API key validation
- API key creation
- API key deactivation
- Hash generation
- Database error handling

**Rate Limit Service (`internal/services/rate_limit_service_test.go`):**
- Rate limit checking
- Rate limit status retrieval
- Edge cases (exactly at limit, over limit)
- Default configuration handling
- Redis error handling

#### Middleware Tests

**Rate Limit Middleware (`internal/middleware/rate_limit_test.go`):**
- API key authentication
- Rate limit enforcement
- Header validation
- Endpoint skipping (health, admin)
- Error responses

**CORS Middleware (`internal/middleware/cors_test.go`):**
- CORS header setting
- Different HTTP methods
- Origin handling

### 2. Integration Tests

Integration tests (`integration_test.go`) test the complete workflow:

- **Complete API workflow**: Create key → Use key → Check limits → Deactivate key
- **Rate limit enforcement**: Test actual rate limiting behavior
- **Error handling**: Test various error scenarios
- **Admin endpoint access**: Verify admin endpoints don't require authentication
- **Health check**: Verify health endpoint works without authentication

## Test Coverage

### Running Coverage Analysis

```bash
# Generate coverage report
make test-coverage

# View coverage in terminal
go test -cover ./...

# View detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Coverage Goals

- **Handlers**: 100% coverage of all endpoints
- **Services**: 100% coverage of all business logic
- **Middleware**: 100% coverage of all middleware functions
- **Overall**: Target 90%+ coverage

## Writing New Tests

### Test Structure

Follow the standard Go testing pattern:

```go
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange
    setup := createTestSetup()
    
    // Act
    result, err := setup.FunctionUnderTest()
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Test Naming Convention

- `TestFunctionName_Success` - Happy path
- `TestFunctionName_Error` - Error cases
- `TestFunctionName_InvalidInput` - Invalid input handling
- `TestFunctionName_EdgeCase` - Edge cases

### Example Test

```go
func TestCreateAPIKey_Success(t *testing.T) {
    // Setup
    router, mockAPIKeyService, _, _ := setupTestRouter()
    mockAPIKeyService.On("CreateAPIKey", "Test Key", 100, 3600).Return("ak_123", nil)
    
    // Create request
    requestBody := map[string]interface{}{
        "name": "Test Key",
        "rate_limit_requests": 100,
        "rate_limit_window_seconds": 3600,
    }
    jsonBody, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", "/admin/api-keys", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    // Execute
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    mockAPIKeyService.AssertExpectations(t)
}
```

## Mocking

### Mock Services

The test suite uses `github.com/stretchr/testify/mock` for mocking:

```go
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
```

### Setting Up Mocks

```go
// Setup expectations
mockService.On("MethodName", expectedArg).Return(expectedResult, nil)

// Verify expectations
mockService.AssertExpectations(t)
```

## Integration Testing

### Test Database Setup

For integration tests, you can use:

1. **In-memory databases** (SQLite)
2. **Test containers** (Docker)
3. **Mock services** (current approach)

### Example Integration Test

```go
func TestIntegration_CompleteWorkflow(t *testing.T) {
    setup := setupIntegrationTest(t)
    
    // Step 1: Create API key
    createRequest := map[string]interface{}{
        "name": "Test Key",
        "rate_limit_requests": 5,
        "rate_limit_window_seconds": 60,
    }
    
    // ... test implementation
}
```

## Continuous Integration

### GitHub Actions

Create `.github/workflows/test.yml`:

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.19
    - run: make test-coverage
```

## Best Practices

### 1. Test Organization
- Group related tests together
- Use descriptive test names
- Keep tests focused and simple

### 2. Mocking
- Mock external dependencies
- Verify mock expectations
- Use realistic test data

### 3. Assertions
- Use specific assertions
- Test both success and failure cases
- Verify error messages

### 4. Test Data
- Use consistent test data
- Create helper functions for test data
- Avoid hardcoded values

### 5. Coverage
- Aim for high coverage
- Focus on critical paths
- Don't test trivial getters/setters

## Troubleshooting

### Common Issues

1. **Import errors**: Run `go mod tidy`
2. **Mock failures**: Check mock expectations
3. **Race conditions**: Use `go test -race`
4. **Slow tests**: Use `go test -short` for unit tests

### Debugging Tests

```bash
# Run specific test
go test -run TestSpecificFunction ./internal/handlers/

# Run with verbose output
go test -v -run TestSpecificFunction ./internal/handlers/

# Run with race detection
go test -race -run TestSpecificFunction ./internal/handlers/
```

## Performance Testing

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkFunctionName ./internal/services/
```

### Load Testing

For load testing, use tools like:
- **Apache Bench (ab)**
- **wrk**
- **Artillery**
- **k6**

Example with Apache Bench:
```bash
# Test rate limiting
ab -n 1000 -c 10 -H "X-API-Key: your-api-key" http://localhost:8080/api/status
```

## Conclusion

This testing suite provides comprehensive coverage of the Rate Limiter API, ensuring reliability and maintainability. The combination of unit tests, integration tests, and proper mocking ensures that all components work correctly both in isolation and together.

For questions or issues with testing, please refer to the Go testing documentation or create an issue in the repository.
