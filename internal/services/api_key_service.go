package services

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"grpc-firstls/internal/database"
)

type APIKeyService struct {
	db *database.DB
}

func NewAPIKeyService(db *database.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) ValidateAPIKey(apiKey string) (*database.APIKey, error) {
	keyHash := s.hashAPIKey(apiKey)
	
	query := `
		SELECT id, key_hash, name, rate_limit_requests, rate_limit_window_seconds, is_active, created_at, updated_at
		FROM api_keys 
		WHERE key_hash = $1 AND is_active = true
	`
	
	var apiKeyRecord database.APIKey
	err := s.db.QueryRow(query, keyHash).Scan(
		&apiKeyRecord.ID,
		&apiKeyRecord.KeyHash,
		&apiKeyRecord.Name,
		&apiKeyRecord.RateLimitRequests,
		&apiKeyRecord.RateLimitWindowSeconds,
		&apiKeyRecord.IsActive,
		&apiKeyRecord.CreatedAt,
		&apiKeyRecord.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	
	return &apiKeyRecord, nil
}

func (s *APIKeyService) CreateAPIKey(name string, rateLimitRequests int, rateLimitWindowSeconds int) (string, error) {
	// Generate a new API key
	apiKey := s.generateAPIKey()
	keyHash := s.hashAPIKey(apiKey)
	
	query := `
		INSERT INTO api_keys (key_hash, name, rate_limit_requests, rate_limit_window_seconds)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	
	var id string
	err := s.db.QueryRow(query, keyHash, name, rateLimitRequests, rateLimitWindowSeconds).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}
	
	return apiKey, nil
}

func (s *APIKeyService) DeactivateAPIKey(apiKey string) error {
	keyHash := s.hashAPIKey(apiKey)
	
	query := `UPDATE api_keys SET is_active = false, updated_at = NOW() WHERE key_hash = $1`
	
	result, err := s.db.Exec(query, keyHash)
	if err != nil {
		return fmt.Errorf("failed to deactivate API key: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	
	return nil
}

func (s *APIKeyService) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", hash)
}

func (s *APIKeyService) generateAPIKey() string {
	// Generate a UUID-based API key
	return fmt.Sprintf("ak_%d_%x", time.Now().Unix(), time.Now().UnixNano())
}
