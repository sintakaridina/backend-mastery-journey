package database

import (
	"time"
)

type APIKey struct {
	ID                    string    `json:"id" db:"id"`
	KeyHash               string    `json:"-" db:"key_hash"`
	Name                  string    `json:"name" db:"name"`
	RateLimitRequests     int       `json:"rate_limit_requests" db:"rate_limit_requests"`
	RateLimitWindowSeconds int      `json:"rate_limit_window_seconds" db:"rate_limit_window_seconds"`
	IsActive              bool      `json:"is_active" db:"is_active"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}
