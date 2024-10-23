package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPasswordResetRequestEncodeToJSON(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	request := PasswordResetRequest{
		Id:        "1",
		UserId:    "1",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH1",
	}

	expected := PasswordResetRequestJSON{
		Id:            request.Id,
		UserId:        request.UserId,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
	}

	var result PasswordResetRequestJSON

	json.Unmarshal([]byte(request.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

func TestPasswordResetRequestEncodeToJSONWithCode(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	code := "12345678"
	request := PasswordResetRequest{
		Id:        "1",
		UserId:    "1",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH1",
	}

	expected := PasswordResetRequestWithCodeJSON{
		Id:            request.Id,
		UserId:        request.UserId,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
		Code:          code,
	}

	var result PasswordResetRequestWithCodeJSON

	json.Unmarshal([]byte(request.EncodeToJSONWithCode(code)), &result)

	assert.Equal(t, expected, result)
}

type PasswordResetRequestJSON struct {
	Id            string `json:"id"`
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
	ExpiresAtUnix int64  `json:"expires_at"`
}

type PasswordResetRequestWithCodeJSON struct {
	Id            string `json:"id"`
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
	ExpiresAtUnix int64  `json:"expires_at"`
	Code          string `json:"code"`
}
