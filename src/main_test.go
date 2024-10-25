package main

import (
	"database/sql"
	"faroe/ratelimit"
	"testing"
	"time"
)

func initializeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db
}

func createEnvironment(db *sql.DB, secret []byte) *Environment {
	env := &Environment{
		db:                                   db,
		secret:                               secret,
		passwordHashingIPRateLimit:           ratelimit.NewTokenBucketRateLimit(5, 10*time.Second),
		loginIPRateLimit:                     ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute),
		createEmailVerificationUserRateLimit: ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute),
		verifyUserEmailRateLimit:             ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute),
		verifyEmailUpdateVerificationCodeLimitCounter: ratelimit.NewLimitCounter(5),
		createPasswordResetIPRateLimit:                ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute),
		verifyPasswordResetCodeLimitCounter:           ratelimit.NewLimitCounter(5),
		totpUserRateLimit:                             ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute),
		recoveryCodeUserRateLimit:                     ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute),
	}
	return env
}

type ErrorJSON struct {
	Error string `json:"error"`
}
