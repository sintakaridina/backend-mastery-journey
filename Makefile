# Makefile for Rate Limiter API

.PHONY: help test test-unit test-integration test-coverage test-verbose build run clean deps

# Default target
help:
	@echo "Available targets:"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run all tests
test: deps
	@echo "Running all tests..."
	go test ./...

# Run unit tests only
test-unit: deps
	@echo "Running unit tests..."
	go test -short ./internal/...

# Run integration tests only
test-integration: deps
	@echo "Running integration tests..."
	go test -run Integration ./...

# Run tests with coverage
test-coverage: deps
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-verbose: deps
	@echo "Running tests with verbose output..."
	go test -v ./...

# Build the application
build: deps
	@echo "Building application..."
	go build -o bin/rate-limiter-api ./cmd/server

# Run the application
run: build
	@echo "Running application..."
	./bin/rate-limiter-api

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run tests for specific package
test-handlers:
	@echo "Running handler tests..."
	go test -v ./internal/handlers/...

test-services:
	@echo "Running service tests..."
	go test -v ./internal/services/...

test-middleware:
	@echo "Running middleware tests..."
	go test -v ./internal/middleware/...

# Run tests with race detection
test-race: deps
	@echo "Running tests with race detection..."
	go test -race ./...

# Run benchmarks
benchmark: deps
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/stretchr/testify@latest
