package repository

import "strings"

// isUniqueViolation checks if a PostgreSQL error is a unique constraint violation.
// This is a heuristic — it checks for the "unique" constraint pattern in the error message.
// A more robust approach would parse the PostgreSQL error code (23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unique") || strings.Contains(msg, "23505")
}

// isForeignKeyViolation checks if a PostgreSQL error is a foreign key violation.
func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23503")
}
