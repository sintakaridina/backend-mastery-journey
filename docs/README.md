# Documentation - Rate Limiter API

This directory contains comprehensive documentation and testing resources for the Rate Limiter API.

## 📁 Files Overview

### 📖 Documentation
- **[POSTMAN_TESTING_GUIDE.md](./POSTMAN_TESTING_GUIDE.md)** - Complete Postman testing guide with step-by-step instructions
- **[README.md](./README.md)** - This file, overview of documentation

### 🧪 Testing Resources
- **[Rate_Limiter_API.postman_collection.json](./Rate_Limiter_API.postman_collection.json)** - Postman collection with all API tests
- **[Rate_Limiter_Local.postman_environment.json](./Rate_Limiter_Local.postman_environment.json)** - Postman environment configuration

## 🚀 Quick Start with Postman

### 1. Import Collection and Environment
1. **Open Postman**
2. **Import Collection**: Click "Import" → Select `Rate_Limiter_API.postman_collection.json`
3. **Import Environment**: Click "Import" → Select `Rate_Limiter_Local.postman_environment.json`
4. **Select Environment**: Choose "Rate Limiter Local" from environment dropdown

### 2. Start API Services
```bash
# In your terminal
docker-compose up -d
```

### 3. Run Tests
1. **Select Collection**: "Rate Limiter API Tests"
2. **Click "Run"** to open Collection Runner
3. **Run All Tests** or select specific tests
4. **View Results** in the test results panel

## 📋 Test Sequence

The collection includes 10 comprehensive tests:

1. **Health Check** - Verify API is running
2. **Create API Key** - Generate authentication token
3. **Test Authentication** - Verify API key works
4. **Check Rate Limit Status** - View rate limit information
5. **Test Endpoint (Rate Limiting)** - Send requests to test rate limiting
6. **Test Rate Limit Exceeded** - Verify HTTP 429 response
7. **Test Invalid API Key** - Test authentication failure
8. **Test Missing API Key** - Test missing authentication
9. **Test with Sample API Key** - Test with pre-configured key
10. **Deactivate API Key** - Clean up test data

## 🔧 Environment Variables

The environment includes these variables:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `base_url` | API base URL | `http://localhost:8080` |
| `api_key` | Dynamic API key (set during testing) | `ak_1703001234_abc123def456` |
| `test_api_key` | Sample API key for quick testing | `hello` |
| `created_api_key` | API key created during testing | `ak_1703001234_abc123def456` |
| `timestamp` | Test timestamp | `1703001234` |
| `test_id` | Unique test ID | `req_123456` |

## 🎯 Testing Scenarios

### Basic Flow Testing
1. **Health Check** → Verify API is accessible
2. **Create API Key** → Generate new authentication token
3. **Test Authentication** → Verify token works
4. **Check Rate Limit** → View current status
5. **Deactivate API Key** → Clean up

### Rate Limiting Testing
1. **Create API Key** with low limit (5 requests/60 seconds)
2. **Send 5 requests** to `/api/test` endpoint
3. **Verify all return 200** status
4. **Send 6th request** → Should return **HTTP 429**
5. **Verify rate limit headers** in all responses

### Error Handling Testing
1. **Test Invalid API Key** → Should return HTTP 401
2. **Test Missing API Key** → Should return HTTP 401
3. **Test Rate Limit Exceeded** → Should return HTTP 429

## 📊 Expected Results

### Successful Test Run
```
✅ Health Check: PASS
✅ API Key Creation: PASS
✅ Authentication: PASS
✅ Rate Limit Status: PASS
✅ Rate Limiting: PASS
✅ Error Handling: PASS
✅ API Key Deactivation: PASS

Total Tests: 10
Passed: 10
Failed: 0
Response Time: < 200ms average
```

### Rate Limit Headers
All responses should include:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in window
- `X-RateLimit-Reset`: Reset time (RFC3339 format)

## 🐛 Troubleshooting

### Common Issues

#### Connection Refused
**Error**: `Error: connect ECONNREFUSED 127.0.0.1:8080`
**Solution**: 
- Verify Docker services: `docker-compose ps`
- Check API health: `curl http://localhost:8080/health`

#### Environment Variables Not Working
**Issue**: `{{base_url}}` not resolving
**Solution**:
- Verify environment is selected in Postman
- Check variable names match exactly
- Ensure variables are set in correct environment

#### Rate Limit Not Working
**Issue**: Requests not being rate limited
**Solution**:
- Check Redis connection: `docker-compose logs redis`
- Verify rate limit configuration
- Check API logs: `docker-compose logs api`

## 🔄 Automation

### Collection Runner
- **Iterations**: 1 (for basic testing)
- **Delay**: 1000ms between requests
- **Data**: None (uses environment variables)

### Test Scripts
Each request includes automated test scripts that verify:
- Response status codes
- Response structure
- Response times
- Rate limit headers
- Error messages

## 📈 Performance Testing

### Load Testing
1. **Create Collection** with rate limit test endpoint
2. **Set up Collection Runner** with multiple iterations
3. **Monitor Response Times** and rate limit behavior
4. **Verify Rate Limiting** works under load

### Benchmarks
- **Response Time**: < 200ms for successful requests
- **Rate Limiting**: Accurate enforcement of limits
- **Error Handling**: Proper HTTP status codes
- **Headers**: All rate limit headers present

## 🎓 Learning Objectives

After completing these tests, you should understand:

1. **API Authentication** - How API key authentication works
2. **Rate Limiting** - How rate limiting is implemented and enforced
3. **Error Handling** - How the API handles various error scenarios
4. **HTTP Status Codes** - Proper use of 200, 401, 429 status codes
5. **Response Headers** - How rate limit information is communicated
6. **Testing Best Practices** - How to test APIs comprehensively

## 📚 Additional Resources

- **[Main README](../README.md)** - Project overview and setup
- **[Rate Limiter README](../RATE_LIMITER_README.md)** - Complete API documentation
- **[Quickstart Guide](../specs/002-title-rate-limiter/quickstart.md)** - Manual testing scenarios
- **[API Specification](../specs/002-title-rate-limiter/contracts/api-spec.yaml)** - OpenAPI specification

## 🤝 Contributing

To improve the testing documentation:

1. **Add new test scenarios** to the collection
2. **Update test scripts** for better validation
3. **Add performance tests** for load testing
4. **Improve error handling** test coverage
5. **Document new features** as they're added

---

**Happy Testing! 🚀**

This documentation provides everything you need to thoroughly test the Rate Limiter API using Postman. Follow the guide for comprehensive validation of all features.
