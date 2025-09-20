# Backend Mastery Journey

This repository contains implementations from the backend mastery course, including various API projects and microservices.

## Current Projects

### Rate Limiter API (001-rate-limiter)

A production-ready rate limiting API built with Go, PostgreSQL, and Redis.

**Features:**
- API key authentication with PostgreSQL storage
- Redis-based rate limiting with configurable windows
- Docker Compose orchestration
- HTTP 429 responses for rate limit exceeded
- Comprehensive documentation and testing

**Quick Start:**
```bash
# Switch to the rate limiter branch
git checkout 001-rate-limiter

# Start all services
docker-compose up -d

# Test the API
curl http://localhost:8080/health
```

**Documentation:**
- [Rate Limiter README](./RATE_LIMITER_README.md) - Complete setup and usage guide
- [Quickstart Guide](./specs/002-title-rate-limiter/quickstart.md) - Testing scenarios
- [API Specification](./specs/002-title-rate-limiter/contracts/api-spec.yaml) - OpenAPI spec

## GitHub Codespaces Setup

### Prerequisites
- GitHub Codespaces enabled
- Docker support in Codespaces

### Running in Codespaces

1. **Open in Codespaces:**
   - Click "Code" → "Codespaces" → "Create codespace"
   - Or use the GitHub CLI: `gh codespace create`

2. **Switch to Rate Limiter Branch:**
   ```bash
   git checkout 001-rate-limiter
   ```

3. **Start the Services:**
   ```bash
   docker-compose up -d
   ```

4. **Verify Services:**
   ```bash
   # Check service status
   docker-compose ps
   
   # Test API health
   curl http://localhost:8080/health
   ```

5. **Test Rate Limiting:**
   ```bash
   # Use the sample API key
   curl -H "X-API-Key: hello" http://localhost:8080/api/status
   ```

### Port Forwarding

In GitHub Codespaces, you may need to forward ports:
- Port 8080: API service
- Port 5432: PostgreSQL (if needed for direct access)
- Port 6379: Redis (if needed for direct access)

### Environment Variables

The project uses environment variables for configuration. Copy `env.example` to `.env` if you need to modify settings:

```bash
cp env.example .env
```

## Project Structure

```
├── cmd/server/                    # Application entry point
├── internal/                      # Internal packages
│   ├── config/                   # Configuration management
│   ├── database/                 # PostgreSQL models & connection
│   ├── redis/                    # Redis client
│   ├── services/                 # Business logic
│   ├── handlers/                 # HTTP handlers
│   └── middleware/               # Middleware (CORS, rate limiting)
├── scripts/                      # Database initialization
├── specs/                        # Feature specifications
├── docker-compose.yml            # Service orchestration
├── Dockerfile                    # API container
└── env.example                   # Environment template
```

## Development

### Local Development

1. **Prerequisites:**
   - Go 1.19+
   - Docker and Docker Compose
   - PostgreSQL and Redis (or use Docker)

2. **Setup:**
   ```bash
   # Install dependencies
   go mod download
   
   # Start dependencies
   docker-compose up -d postgres redis
   
   # Run the application
   go run cmd/server/main.go
   ```

### Testing

```bash
# Run all tests
go test ./...

# Test with Docker
docker-compose up -d
curl http://localhost:8080/health
```

## Contributing

1. Create a feature branch
2. Implement your changes
3. Add tests and documentation
4. Submit a pull request

## License

This project is part of the backend mastery course materials.