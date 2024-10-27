package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanUpDatabase(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user1 := User{
		Id:             "1",
		CreatedAt:      now,
		PasswordHash:   "HASH",
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user1)
	if err != nil {
		t.Fatal(err)
	}

	resetRequest1 := PasswordResetRequest{
		Id:        "1",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(-10 * time.Minute),
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
	if err != nil {
		t.Fatal(err)
	}

	resetRequest2 := PasswordResetRequest{
		Id:        "2",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
	if err != nil {
		t.Fatal(err)
	}

	resetRequest3 := PasswordResetRequest{
		Id:        "3",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &resetRequest3)
	if err != nil {
		t.Fatal(err)
	}

	user2 := User{
		Id:             "2",
		CreatedAt:      now,
		PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	user3 := User{
		Id:             "3",
		CreatedAt:      now,
		PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ",
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user3)
	if err != nil {
		t.Fatal(err)
	}

	verificationRequest1 := UserEmailVerificationRequest{
		UserId:    user1.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertUserEmailVerificationRequest(db, &verificationRequest1)
	if err != nil {
		t.Fatal(err)
	}

	verificationRequest2 := UserEmailVerificationRequest{
		UserId:    user2.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(-10 * time.Minute),
	}
	err = insertUserEmailVerificationRequest(db, &verificationRequest2)
	if err != nil {
		t.Fatal(err)
	}

	verificationRequest3 := UserEmailVerificationRequest{
		UserId:    user3.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(-10 * time.Minute),
	}
	err = insertUserEmailVerificationRequest(db, &verificationRequest3)
	if err != nil {
		t.Fatal(err)
	}

	err = cleanUpDatabase(db)
	if err != nil {
		t.Fatal(err)
	}

	var passwordResetRequestCount int
	err = db.QueryRow("SELECT count(*) FROM password_reset_request").Scan(&passwordResetRequestCount)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, passwordResetRequestCount)

	var emailVerificationRequestCount int
	err = db.QueryRow("SELECT count(*) FROM user_email_verification_request").Scan(&emailVerificationRequestCount)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, emailVerificationRequestCount)
}
