package services

import (
	"database/sql"
	"testing"
	"time"

	"grpc-firstls/internal/database"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Helper function to create test API key data

func createTestAPIKeyForAPIKeyService() *database.APIKey {
	return &database.APIKey{
		ID:                     "test-id-123",
		KeyHash:                "test-hash-abc123",
		Name:                   "Test API Key",
		RateLimitRequests:      100,
		RateLimitWindowSeconds: 3600,
		IsActive:               true,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}

func TestAPIKeyService_ValidateAPIKey_Success(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Create test data
	testAPIKey := "ak_1234567890_abcdef"
	expectedAPIKey := createTestAPIKeyForAPIKeyService()
	expectedHash := service.hashAPIKey(testAPIKey)

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{"id", "key_hash", "name", "rate_limit_requests", "rate_limit_window_seconds", "is_active", "created_at", "updated_at"}).
		AddRow(expectedAPIKey.ID, expectedAPIKey.KeyHash, expectedAPIKey.Name, expectedAPIKey.RateLimitRequests, expectedAPIKey.RateLimitWindowSeconds, expectedAPIKey.IsActive, expectedAPIKey.CreatedAt, expectedAPIKey.UpdatedAt)

	mock.ExpectQuery(`SELECT id, key_hash, name, rate_limit_requests, rate_limit_window_seconds, is_active, created_at, updated_at`).
		WithArgs(expectedHash).
		WillReturnRows(rows)

	// Call the method
	result, err := service.ValidateAPIKey(testAPIKey)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAPIKey.ID, result.ID)
	assert.Equal(t, expectedAPIKey.Name, result.Name)
	assert.Equal(t, expectedAPIKey.RateLimitRequests, result.RateLimitRequests)
	assert.Equal(t, expectedAPIKey.RateLimitWindowSeconds, result.RateLimitWindowSeconds)
	assert.True(t, result.IsActive)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_ValidateAPIKey_NotFound(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Create test data
	testAPIKey := "invalid-key"
	expectedHash := service.hashAPIKey(testAPIKey)

	// Setup mock expectations - return sql.ErrNoRows
	mock.ExpectQuery(`SELECT id, key_hash, name, rate_limit_requests, rate_limit_window_seconds, is_active, created_at, updated_at`).
		WithArgs(expectedHash).
		WillReturnError(sql.ErrNoRows)

	// Call the method
	result, err := service.ValidateAPIKey(testAPIKey)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid API key")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_ValidateAPIKey_DatabaseError(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Create test data
	testAPIKey := "test-key"
	expectedHash := service.hashAPIKey(testAPIKey)

	// Setup mock expectations - return database error
	mock.ExpectQuery(`SELECT id, key_hash, name, rate_limit_requests, rate_limit_window_seconds, is_active, created_at, updated_at`).
		WithArgs(expectedHash).
		WillReturnError(assert.AnError)

	// Call the method
	result, err := service.ValidateAPIKey(testAPIKey)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to validate API key")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_CreateAPIKey_Success(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{"id"}).AddRow("new-id-123")

	mock.ExpectQuery(`INSERT INTO api_keys`).
		WithArgs(sqlmock.AnyArg(), "Test API Key", 100, 3600).
		WillReturnRows(rows)

	// Call the method
	apiKey, err := service.CreateAPIKey("Test API Key", 100, 3600)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, apiKey)
	assert.Contains(t, apiKey, "ak_") // Should start with "ak_"

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_CreateAPIKey_DatabaseError(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Setup mock expectations - return database error
	mock.ExpectQuery(`INSERT INTO api_keys`).
		WithArgs(sqlmock.AnyArg(), "Test API Key", 100, 3600).
		WillReturnError(assert.AnError)

	// Call the method
	apiKey, err := service.CreateAPIKey("Test API Key", 100, 3600)

	// Assertions
	assert.Error(t, err)
	assert.Empty(t, apiKey)
	assert.Contains(t, err.Error(), "failed to create API key")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_DeactivateAPIKey_Success(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Setup mock expectations
	mock.ExpectExec(`UPDATE api_keys SET is_active = false, updated_at = NOW\(\) WHERE key_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method
	err = service.DeactivateAPIKey("test-api-key")

	// Assertions
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_DeactivateAPIKey_NotFound(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Setup mock expectations - no rows affected
	mock.ExpectExec(`UPDATE api_keys SET is_active = false, updated_at = NOW\(\) WHERE key_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method
	err = service.DeactivateAPIKey("non-existent-key")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not found")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_DeactivateAPIKey_DatabaseError(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Setup mock expectations - return database error
	mock.ExpectExec(`UPDATE api_keys SET is_active = false, updated_at = NOW\(\) WHERE key_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(assert.AnError)

	// Call the method
	err = service.DeactivateAPIKey("test-api-key")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deactivate API key")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_DeactivateAPIKey_RowsAffectedError(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Setup mock expectations - error getting rows affected
	mock.ExpectExec(`UPDATE api_keys SET is_active = false, updated_at = NOW\(\) WHERE key_hash = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(assert.AnError))

	// Call the method
	err = service.DeactivateAPIKey("test-api-key")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get rows affected")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_hashAPIKey(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Test that the same input produces the same hash
	apiKey := "test-api-key-123"
	hash1 := service.hashAPIKey(apiKey)
	hash2 := service.hashAPIKey(apiKey)

	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, apiKey, hash1) // Hash should be different from original
	assert.Len(t, hash1, 64)          // SHA256 produces 64 character hex string

	// Test that different inputs produce different hashes
	differentKey := "different-api-key-456"
	hash3 := service.hashAPIKey(differentKey)

	assert.NotEqual(t, hash1, hash3)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyService_generateAPIKey(t *testing.T) {
	// Create a real database connection with sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create service with real database connection
	service := NewAPIKeyService(db)

	// Generate multiple API keys
	key1 := service.generateAPIKey()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	key2 := service.generateAPIKey()

	// Keys should be different
	assert.NotEqual(t, key1, key2)

	// Keys should start with "ak_"
	assert.True(t, len(key1) > 3)
	assert.True(t, len(key2) > 3)
	assert.Contains(t, key1, "ak_")
	assert.Contains(t, key2, "ak_")

	// Keys should contain timestamp information
	assert.Contains(t, key1, "_")
	assert.Contains(t, key2, "_")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
