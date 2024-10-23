package main

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func insertUserEmailVerificationRequest(db *sql.DB, request *UserEmailVerificationRequest) error {
	_, err := db.Exec("INSERT INTO user_email_verification_request (user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?)", request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code, request.CreatedAt.Unix(), request.Code, request.UserId)
	return err
}
func TestEncodeEmailToJSON(t *testing.T) {
	t.Parallel()

	email := "user@example.com"

	expected := EmailJSON{
		Email: email,
	}

	var result EmailJSON

	json.Unmarshal([]byte(encodeEmailToJSON(email)), &result)

	assert.Equal(t, expected, result)
}

func TestEmailUpdateRequestEncodeToJSON(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	request := EmailUpdateRequest{
		Id:        "1",
		UserId:    "1",
		Email:     "user@example.com",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		Code:      "12345678",
	}

	expected := EmailUpdateRequestJSON{
		Id:            request.Id,
		UserId:        request.UserId,
		Email:         request.Email,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
		Code:          request.Code,
	}

	var result EmailUpdateRequestJSON

	json.Unmarshal([]byte(request.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

func TestUserEmailVerificationRequestEncodeToJSON(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	request := UserEmailVerificationRequest{
		UserId:    "1",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		Code:      "12345678",
	}

	expected := UserEmailVerificationRequestJSON{
		UserId:        request.UserId,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
		Code:          request.Code,
	}

	var result UserEmailVerificationRequestJSON

	json.Unmarshal([]byte(request.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

type EmailJSON struct {
	Email string `json:"email"`
}

type EmailUpdateRequestJSON struct {
	Id            string `json:"id"`
	UserId        string `json:"user_id"`
	Email         string `json:"email"`
	CreatedAtUnix int64  `json:"created_at"`
	ExpiresAtUnix int64  `json:"expires_at"`
	Code          string `json:"code"`
}

type UserEmailVerificationRequestJSON struct {
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
	ExpiresAtUnix int64  `json:"expires_at"`
	Code          string `json:"code"`
}
