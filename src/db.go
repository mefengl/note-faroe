package main

import (
	"database/sql"
	"time"
)

func cleanUpDatabase(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM user_email_verification_request WHERE expires_at <= ?", time.Now().Unix())
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM password_reset_request WHERE expires_at <= ?", time.Now().Unix())
	if err != nil {
		return err
	}
	return err
}
