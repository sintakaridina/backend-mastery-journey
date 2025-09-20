# Stage 1: Build
FROM golang:1.21 AS builder

# Set working directory inside container
WORKDIR /app

# Copy dependency file and download dependency
COPY go.mod ./
RUN go mod download

# Copy all source code to container
COPY . .

# Build binary
RUN go build -o app main.go

# Stage 2: Run (lightweight image)
FROM debian:bookworm-slim

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/app .

# Expose port (change if your application listen on other port)
EXPOSE 8080

# Run application
CMD ["./app"]
