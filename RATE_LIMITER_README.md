# Rate Limiter API

A production-ready rate limiting API built with Go, PostgreSQL, and Redis. This service provides API key authentication and configurable rate limiting for protecting your endpoints.

## Features

- **API Key Authentication**: Secure API key-based authentication with PostgreSQL storage
- **Rate Limiting**: Configurable rate limits per API key using Redis for fast access
- **HTTP 429 Responses**: Proper rate limit exceeded responses with retry information
- **Docker Support**: Complete Docker Compose setup for easy deployment
- **Production Ready**: Health checks, proper error handling, and monitoring headers

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───▶│  Rate Limit │───▶│ PostgreSQL  │
│             │    │     API     │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │    Redis    │
                   │             │
                   └─────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Using Docker Compose (Recommended)

1. **Clone and navigate to the project**:
   ```bash
   cd grpc-firstls
   ```

2. **Start all services**:
   ```bash
   docker-compose up -d
   ```

3. **Verify services are running**:
   ```bash
   docker-compose ps
   ```

4. **Check API health**:
   ```bash
   curl http://localhost:8080/health
   ```

### Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Start PostgreSQL and Redis**:
   ```bash
   docker-compose up -d postgres redis
   ```

3. **Copy environment file**:
   ```bash
   cp env.example .env
   ```

4. **Run the application**:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

### Health Check
```http
GET /health
```
Returns the service health status (no authentication required).

### Create API Key
```http
POST /admin/api-keys
Content-Type: application/json

{
  "name": "My API Key",
  "rate_limit_requests": 100,
  "rate_limit_window_seconds": 3600
}
```

### Deactivate API Key
```http
DELETE /admin/api-keys/{api_key}
```

### Protected Endpoints

All endpoints below require authentication via `X-API-Key` header or `Authorization: Bearer {api_key}` header.

#### Get Status
```http
GET /api/status
X-API-Key: your-api-key-here
```

#### Get Rate Limit Status
```http
GET /api/rate-limit
X-API-Key: your-api-key-here
```

#### Test Endpoint
```http
POST /api/test
X-API-Key: your-api-key-here
Content-Type: application/json

{
  "message": "Hello, World!"
}
```

## Rate Limiting

### How It Works

1. **Authentication**: Each request is validated against PostgreSQL-stored API keys
2. **Rate Limiting**: Redis tracks request counts per API key with configurable windows
3. **Headers**: Rate limit information is included in response headers:
   - `X-RateLimit-Limit`: Maximum requests allowed
   - `X-RateLimit-Remaining`: Requests remaining in current window
   - `X-RateLimit-Reset`: When the rate limit window resets

### Rate Limit Responses

When rate limit is exceeded:
```json
{
  "error": "Rate limit exceeded",
  "message": "You have exceeded your rate limit. Please try again later.",
  "retry_after": 3600
}
```

HTTP Status: `429 Too Many Requests`

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:password@localhost:5432/rate_limiter?sslmode=disable` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `PORT` | `8080` | Server port |
| `DEFAULT_RATE_LIMIT_REQUESTS` | `100` | Default requests per window |
| `DEFAULT_RATE_LIMIT_WINDOW` | `1h` | Default time window |
| `GIN_MODE` | `release` | Gin framework mode |

### Database Schema

The API uses a single table for API key management:

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    rate_limit_requests INTEGER NOT NULL DEFAULT 100,
    rate_limit_window_seconds INTEGER NOT NULL DEFAULT 3600,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Testing

### Create a Test API Key

```bash
curl -X POST http://localhost:8080/admin/api-keys \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Key",
    "rate_limit_requests": 5,
    "rate_limit_window_seconds": 60
  }'
```

### Test Rate Limiting

```bash
# Replace YOUR_API_KEY with the key from the previous response
API_KEY="your-api-key-here"

# Make requests to test rate limiting
for i in {1..7}; do
  echo "Request $i:"
  curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/status
  echo -e "\n"
  sleep 1
done
```

### Test with Sample API Key

The database is initialized with a test API key:
- **API Key**: `hello`
- **Rate Limit**: 10 requests per 60 seconds

```bash
# Test with the sample key
curl -H "X-API-Key: hello" http://localhost:8080/api/status
```

## Development

### Project Structure

```
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── database/
│   │   ├── database.go         # Database connection
│   │   └── models.go           # Data models
│   ├── handlers/
│   │   └── handlers.go         # HTTP handlers
│   ├── middleware/
│   │   ├── cors.go             # CORS middleware
│   │   └── rate_limit.go       # Rate limiting middleware
│   ├── redis/
│   │   └── redis.go            # Redis client
│   └── services/
│       ├── api_key_service.go  # API key management
│       └── rate_limit_service.go # Rate limiting logic
├── scripts/
│   └── init-db.sql             # Database initialization
├── docker-compose.yml          # Docker services
├── Dockerfile                  # API container
└── env.example                 # Environment template
```

### Adding New Endpoints

1. **Add handler method** in `internal/handlers/handlers.go`
2. **Register route** in `SetupRoutes()` method
3. **Protected endpoints** automatically get rate limiting via middleware

### Customizing Rate Limits

Rate limits can be configured per API key or globally:

- **Per API Key**: Set `rate_limit_requests` and `rate_limit_window_seconds` when creating the key
- **Global Defaults**: Modify `DEFAULT_RATE_LIMIT_REQUESTS` and `DEFAULT_RATE_LIMIT_WINDOW` environment variables

## Monitoring

### Health Checks

- **API Health**: `GET /health`
- **Docker Health**: Built-in health checks for all services

### Logs

View logs for all services:
```bash
docker-compose logs -f
```

View logs for specific service:
```bash
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f redis
```

## Production Considerations

1. **Security**: 
   - Use strong API keys
   - Enable HTTPS in production
   - Consider API key rotation

2. **Performance**:
   - Monitor Redis memory usage
   - Consider Redis clustering for high load
   - Database connection pooling

3. **Monitoring**:
   - Add metrics collection (Prometheus)
   - Set up alerting for rate limit violations
   - Monitor database and Redis performance

4. **Scaling**:
   - Use load balancers for multiple API instances
   - Consider Redis Sentinel for high availability
   - Database read replicas for read-heavy workloads

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure all services are running with `docker-compose ps`
2. **Database Connection Failed**: Check PostgreSQL logs with `docker-compose logs postgres`
3. **Redis Connection Failed**: Check Redis logs with `docker-compose logs redis`
4. **Rate Limit Not Working**: Verify Redis is accessible and check API key configuration

### Debug Mode

Set `GIN_MODE=debug` in your environment for detailed request logging.

## License

This project is part of the backend-mastery course materials.
