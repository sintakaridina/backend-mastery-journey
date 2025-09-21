apa y# Postman Testing Guide - Rate Limiter API

**API**: Rate Limiter API  
**Version**: 1.0.0  
**Base URL**: `http://localhost:8080`  
**Date**: 2024-12-19

## 📋 Table of Contents

1. [Setup Postman](#setup-postman)
2. [Environment Configuration](#environment-configuration)
3. [API Testing Flow](#api-testing-flow)
4. [Individual Endpoint Tests](#individual-endpoint-tests)
5. [Rate Limiting Tests](#rate-limiting-tests)
6. [Error Handling Tests](#error-handling-tests)
7. [Automation Scripts](#automation-scripts)
8. [Troubleshooting](#troubleshooting)

## 🚀 Setup Postman

### Prerequisites
- Postman Desktop App or Web Version
- Rate Limiter API running locally (`docker-compose up -d`)
- Docker services accessible on `localhost:8080`

### Collection Setup
1. **Create New Collection**: "Rate Limiter API Tests"
2. **Add Description**: "Comprehensive testing for Rate Limiter API with authentication and rate limiting"
3. **Set Collection Variables**:
   ```
   base_url: http://localhost:8080
   api_key: (will be set after creating API key)
   ```

## 🔧 Environment Configuration

### Create Environment: "Rate Limiter Local"
```json
{
  "base_url": "http://localhost:8080",
  "api_key": "",
  "test_api_key": "hello",
  "created_api_key": ""
}
```

### Environment Variables Usage
- `{{base_url}}` - Base URL for all requests
- `{{api_key}}` - Dynamic API key from creation
- `{{test_api_key}}` - Sample API key for quick testing
- `{{created_api_key}}` - API key created during testing

## 🔄 API Testing Flow

### Complete Testing Sequence
1. **Health Check** → Verify API is running
2. **Create API Key** → Generate new authentication token
3. **Test Authentication** → Verify API key works
4. **Check Rate Limit Status** → View current rate limit info
5. **Test Rate Limiting** → Send multiple requests to trigger rate limit
6. **Verify Rate Limit Response** → Confirm HTTP 429 response
7. **Deactivate API Key** → Clean up test data

## 📡 Individual Endpoint Tests

### 1. Health Check
**Purpose**: Verify API service is running and accessible

**Request**:
```
Method: GET
URL: {{base_url}}/health
Headers: None
Body: None
```

**Expected Response**:
```json
{
  "status": "healthy",
  "service": "rate-limiter-api"
}
```

**Test Script**:
```javascript
pm.test("Health check returns 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response contains service status", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.status).to.eql("healthy");
    pm.expect(jsonData.service).to.eql("rate-limiter-api");
});
```

### 2. Create API Key
**Purpose**: Generate new API key for authentication

**Request**:
```
Method: POST
URL: {{base_url}}/admin/api-keys
Headers: 
  Content-Type: application/json
Body (raw JSON):
{
  "name": "Postman Test Key",
  "rate_limit_requests": 5,
  "rate_limit_window_seconds": 60
}
```

**Expected Response**:
```json
{
  "api_key": "ak_1703001234_abc123def456",
  "name": "Postman Test Key",
  "rate_limit": {
    "requests": 5,
    "window_seconds": 60
  }
}
```

**Test Script**:
```javascript
pm.test("API key creation returns 201", function () {
    pm.response.to.have.status(201);
});

pm.test("Response contains API key", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.api_key).to.exist;
    pm.expect(jsonData.name).to.eql("Postman Test Key");
    
    // Set environment variable for subsequent requests
    pm.environment.set("api_key", jsonData.api_key);
    pm.environment.set("created_api_key", jsonData.api_key);
});
```

### 3. Test Authentication
**Purpose**: Verify API key authentication works

**Request**:
```
Method: GET
URL: {{base_url}}/api/status
Headers: 
  X-API-Key: {{api_key}}
Body: None
```

**Expected Response**:
```json
{
  "status": "authenticated",
  "api_key": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Postman Test Key"
  }
}
```

**Test Script**:
```javascript
pm.test("Authentication returns 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response indicates authenticated status", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.status).to.eql("authenticated");
    pm.expect(jsonData.api_key).to.exist;
});
```

### 4. Check Rate Limit Status
**Purpose**: View current rate limit information

**Request**:
```
Method: GET
URL: {{base_url}}/api/rate-limit
Headers: 
  X-API-Key: {{api_key}}
Body: None
```

**Expected Response**:
```json
{
  "rate_limit": {
    "limit": 5,
    "remaining": 5,
    "reset_time": "2024-12-19T15:30:00Z",
    "allowed": true
  }
}
```

**Test Script**:
```javascript
pm.test("Rate limit status returns 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Rate limit headers present", function () {
    pm.expect(pm.response.headers.get("X-RateLimit-Limit")).to.exist;
    pm.expect(pm.response.headers.get("X-RateLimit-Remaining")).to.exist;
    pm.expect(pm.response.headers.get("X-RateLimit-Reset")).to.exist;
});

pm.test("Rate limit data is valid", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.rate_limit.limit).to.be.a('number');
    pm.expect(jsonData.rate_limit.remaining).to.be.a('number');
    pm.expect(jsonData.rate_limit.allowed).to.be.a('boolean');
});
```

### 5. Test Endpoint (Rate Limiting)
**Purpose**: Send requests to test rate limiting behavior

**Request**:
```
Method: POST
URL: {{base_url}}/api/test
Headers: 
  X-API-Key: {{api_key}}
  Content-Type: application/json
Body (raw JSON):
{
  "message": "Test message from Postman"
}
```

**Expected Response (Success)**:
```json
{
  "message": "Request processed successfully",
  "echo": "Test message from Postman",
  "api_key": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Postman Test Key"
  }
}
```

**Test Script**:
```javascript
pm.test("Request processed successfully", function () {
    pm.response.to.have.status(200);
});

pm.test("Response contains echo message", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.message).to.eql("Request processed successfully");
    pm.expect(jsonData.echo).to.eql("Test message from Postman");
});

pm.test("Response time is acceptable", function () {
    pm.expect(pm.response.responseTime).to.be.below(200);
});
```

### 6. Deactivate API Key
**Purpose**: Clean up test data by deactivating API key

**Request**:
```
Method: DELETE
URL: {{base_url}}/admin/api-keys/{{created_api_key}}
Headers: None
Body: None
```

**Expected Response**:
```json
{
  "message": "API key deactivated successfully"
}
```

**Test Script**:
```javascript
pm.test("API key deactivation returns 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Deactivation message is correct", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.message).to.eql("API key deactivated successfully");
});
```

## ⚡ Rate Limiting Tests

### Quick Test with Sample API Key
**API Key**: `hello` (pre-configured)
**Rate Limit**: 10 requests per 60 seconds

**Test Sequence**:
1. **Send 10 requests** to `/api/test` with `X-API-Key: hello`
2. **Verify all return 200** status
3. **Send 11th request** - should return **HTTP 429**
4. **Verify rate limit headers** in all responses

**Request Template**:
```
Method: POST
URL: {{base_url}}/api/test
Headers: 
  X-API-Key: hello
  Content-Type: application/json
Body (raw JSON):
{
  "message": "Rate limit test {{$randomInt}}"
}
```

### Rate Limit Exceeded Test
**Expected Response (HTTP 429)**:
```json
{
  "error": "Rate limit exceeded",
  "message": "You have exceeded your rate limit. Please try again later.",
  "retry_after": 60
}
```

**Test Script for 429 Response**:
```javascript
pm.test("Rate limit exceeded returns 429", function () {
    pm.response.to.have.status(429);
});

pm.test("Error message is correct", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.error).to.eql("Rate limit exceeded");
    pm.expect(jsonData.retry_after).to.be.a('number');
});
```

## ❌ Error Handling Tests

### 1. Invalid API Key Test
**Request**:
```
Method: GET
URL: {{base_url}}/api/status
Headers: 
  X-API-Key: invalid-key-123
Body: None
```

**Expected Response (HTTP 401)**:
```json
{
  "error": "Invalid API key",
  "message": "The provided API key is invalid or inactive"
}
```

### 2. Missing API Key Test
**Request**:
```
Method: GET
URL: {{base_url}}/api/status
Headers: None
Body: None
```

**Expected Response (HTTP 401)**:
```json
{
  "error": "API key required",
  "message": "Please provide an API key in the X-API-Key header or Authorization header"
}
```

### 3. Invalid Request Body Test
**Request**:
```
Method: POST
URL: {{base_url}}/api/test
Headers: 
  X-API-Key: {{api_key}}
  Content-Type: application/json
Body (raw JSON):
{
  "invalid_field": "test"
}
```

**Expected Response (HTTP 400)**:
```json
{
  "error": "Invalid request",
  "message": "Request validation failed"
}
```

## 🤖 Automation Scripts

### Collection-Level Pre-request Script
```javascript
// Set timestamp for unique test data
pm.environment.set("timestamp", new Date().getTime());
pm.environment.set("test_id", pm.info.requestId);
```

### Collection-Level Test Script
```javascript
// Global response time check
pm.test("Response time is acceptable", function () {
    pm.expect(pm.response.responseTime).to.be.below(500);
});

// Global error handling
if (pm.response.code >= 400) {
    pm.test("Error response has proper structure", function () {
        const jsonData = pm.response.json();
        pm.expect(jsonData.error).to.exist;
        pm.expect(jsonData.message).to.exist;
    });
}
```

### Rate Limit Monitoring Script
```javascript
// Check rate limit headers on every response
pm.test("Rate limit headers present", function () {
    if (pm.response.code === 200 || pm.response.code === 429) {
        pm.expect(pm.response.headers.get("X-RateLimit-Limit")).to.exist;
        pm.expect(pm.response.headers.get("X-RateLimit-Remaining")).to.exist;
        pm.expect(pm.response.headers.get("X-RateLimit-Reset")).to.exist;
    }
});

// Log rate limit information
if (pm.response.headers.get("X-RateLimit-Limit")) {
    console.log("Rate Limit Info:");
    console.log("Limit:", pm.response.headers.get("X-RateLimit-Limit"));
    console.log("Remaining:", pm.response.headers.get("X-RateLimit-Remaining"));
    console.log("Reset:", pm.response.headers.get("X-RateLimit-Reset"));
}
```

## 🔄 Postman Collection Runner

### Test Sequence for Collection Runner
1. **Health Check** (1 iteration)
2. **Create API Key** (1 iteration)
3. **Test Authentication** (1 iteration)
4. **Check Rate Limit Status** (1 iteration)
5. **Test Endpoint** (6 iterations - to trigger rate limit)
6. **Deactivate API Key** (1 iteration)

### Runner Configuration
```
Iterations: 1
Delay: 1000ms between requests
Data: None (use environment variables)
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Connection Refused
**Error**: `Error: connect ECONNREFUSED 127.0.0.1:8080`
**Solution**: 
- Verify Docker services are running: `docker-compose ps`
- Check API health: `curl http://localhost:8080/health`

#### 2. Invalid API Key
**Error**: HTTP 401 with "Invalid API key"
**Solution**:
- Verify API key is correctly set in environment
- Check if API key was created successfully
- Ensure API key is not deactivated

#### 3. Rate Limit Not Working
**Issue**: Requests not being rate limited
**Solution**:
- Check Redis connection: `docker-compose logs redis`
- Verify rate limit configuration in API key creation
- Check API logs: `docker-compose logs api`

#### 4. Environment Variables Not Working
**Issue**: `{{base_url}}` not resolving
**Solution**:
- Verify environment is selected in Postman
- Check environment variable names match exactly
- Ensure variables are set in correct environment

### Debug Information
```javascript
// Add to test scripts for debugging
console.log("Request URL:", pm.request.url);
console.log("Request Headers:", pm.request.headers);
console.log("Response Status:", pm.response.status);
console.log("Response Time:", pm.response.responseTime);
console.log("Environment Variables:", pm.environment.toObject());
```

## 📊 Performance Testing

### Load Testing with Postman
1. **Create Collection** with rate limit test endpoint
2. **Set up Collection Runner** with multiple iterations
3. **Monitor Response Times** and rate limit behavior
4. **Verify Rate Limiting** works under load

### Performance Benchmarks
- **Response Time**: < 200ms for successful requests
- **Rate Limiting**: Accurate enforcement of limits
- **Error Handling**: Proper HTTP status codes
- **Headers**: All rate limit headers present

## 📝 Test Report Template

### Test Results Summary
```
✅ Health Check: PASS
✅ API Key Creation: PASS
✅ Authentication: PASS
✅ Rate Limit Status: PASS
✅ Rate Limiting: PASS
✅ Error Handling: PASS
✅ API Key Deactivation: PASS

Total Tests: 7
Passed: 7
Failed: 0
Response Time: < 200ms average
```

## 🎯 Best Practices

1. **Use Environment Variables** for dynamic data
2. **Set up Test Scripts** for automated validation
3. **Monitor Response Times** for performance
4. **Test Error Scenarios** thoroughly
5. **Clean up Test Data** after testing
6. **Document Test Results** for future reference

---

**Happy Testing! 🚀**

This guide provides comprehensive testing coverage for the Rate Limiter API using Postman. Follow the sequence for complete validation of all features.
