package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserEncodeToJSON(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	user := User{
		Id:             "1",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}

	expected := UserJSON{
		Id:             user.Id,
		CreatedAtUnix:  user.CreatedAt.Unix(),
		TOTPRegistered: user.TOTPRegistered,
		RecoveryCode:   user.RecoveryCode,
	}

	var result UserJSON

	json.Unmarshal([]byte(user.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

func TestEncodeRecoveryCodeToJSON(t *testing.T) {
	t.Parallel()

	recoveryCode := "12345678"

	expected := RecoveryCodeJSON{
		RecoveryCode: recoveryCode,
	}

	var result RecoveryCodeJSON

	json.Unmarshal([]byte(encodeRecoveryCodeToJSON(recoveryCode)), &result)

	assert.Equal(t, expected, result)
}

type UserJSON struct {
	Id             string `json:"id"`
	CreatedAtUnix  int64  `json:"created_at"`
	RecoveryCode   string `json:"recovery_code"`
	TOTPRegistered bool   `json:"totp_registered"`
}

type RecoveryCodeJSON struct {
	RecoveryCode string `json:"recovery_code"`
}
