package redis

import (
	"context"
	"time"
)

// ClientInterface defines the interface for Redis operations
type ClientInterface interface {
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)
	GetRateLimitCount(ctx context.Context, key string) (int64, error)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
