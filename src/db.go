package main

import (
	"database/sql" // Provides generic interface around SQL (or SQL-like) databases.
	"time"         // Provides functionality for measuring and displaying time.
)

// cleanUpDatabase performs routine cleanup tasks on the database.
// Currently, it focuses on removing expired records from request tables
// to prevent them from accumulating indefinitely.
//
// Parameters:
//   db (*sql.DB): A pointer to the active database connection pool.
//
// Returns:
//   error: An error if any of the database delete operations fail, otherwise nil.
//
// How it works:
// 1. It executes a DELETE statement on the 'user_email_verification_request' table.
//    It removes all rows where the 'expires_at' timestamp is less than or equal to
//    the current Unix timestamp (obtained via time.Now().Unix()).
// 2. It checks for errors after the first DELETE operation. If an error occurred,
//    it returns the error immediately.
// 3. If the first operation was successful, it executes a similar DELETE statement
//    on the 'password_reset_request' table, removing expired password reset requests.
// 4. It returns any error that occurred during the second DELETE operation, or nil
//    if both operations were successful.
//
// Usage:
// This function should be called periodically (e.g., on server startup or via a
// scheduled background task) to maintain the database hygiene.
func cleanUpDatabase(db *sql.DB) error {
	// Delete expired email verification requests.
	_, err := db.Exec("DELETE FROM user_email_verification_request WHERE expires_at <= ?", time.Now().Unix())
	if err != nil {
		// If an error occurs, return it immediately.
		return err
	}

	// Delete expired password reset requests.
	_, err = db.Exec("DELETE FROM password_reset_request WHERE expires_at <= ?", time.Now().Unix())
	if err != nil {
		// If an error occurs here, return it.
		return err
	}

	// Return nil if both delete operations were successful.
	// Note: The original code returned 'err' here, which would be nil if the second
	// operation succeeded. Returning nil explicitly is slightly clearer.
	return nil
}
