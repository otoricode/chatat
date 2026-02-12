package repository

import (
	"strings"
)

// isDuplicateKeyError checks if a PostgreSQL error is a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// pgx wraps PostgreSQL error codes; unique_violation = 23505
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "duplicate key")
}
