package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func backupDatabase() error {
	_, err := db.Exec("BEGIN IMMEDIATE")
	if err != nil {
		return err
	}
	src, err := os.Open("faroe_data/sqlite.db")
	if err != nil {
		db.Exec("COMMIT")
		return err
	}
	dst, err := os.Create(fmt.Sprintf("faroe_data/backups/%d.db", time.Now().Unix()))
	if err != nil {
		src.Close()
		db.Exec("COMMIT")
		return err
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		src.Close()
		dst.Close()
		db.Exec("COMMIT")
		return err
	}
	src.Close()
	dst.Close()
	_, err = db.Exec("COMMIT")
	return err
}

func cleanUpDatabase() error {
	_, err := db.Exec("DELETE FROM email_verification_request WHERE expires_at <= ?", time.Now().Unix())
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM password_reset_request WHERE expires_at <= ?", time.Now().Unix())
	if err != nil {
		return err
	}
	return err
}
