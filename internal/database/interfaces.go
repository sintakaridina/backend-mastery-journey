package database

import "database/sql"

// DBInterface defines the interface for database operations
type DBInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Close() error
	Ping() error
}

// Ensure DB implements DBInterface
var _ DBInterface = (*DB)(nil)
