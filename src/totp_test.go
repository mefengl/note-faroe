package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func insertUserTOTPCredential(db *sql.DB, credential *UserTOTPCredential) error {
	_, err := db.Exec("INSERT INTO user_totp_credential (user_id, created_at, key) VALUES (?, ?, ?)", credential.UserId, credential.CreatedAt.Unix(), credential.Key)
	return err
}

func TestUserTOTPCredentialEncodeToJSON(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	credential := UserTOTPCredential{
		UserId:    "1",
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}

	expected := UserTOTPCredentialJSON{
		UserId:        credential.UserId,
		CreatedAtUnix: credential.CreatedAt.Unix(),
		EncodedKey:    base64.StdEncoding.EncodeToString(credential.Key),
	}

	var result UserTOTPCredentialJSON

	json.Unmarshal([]byte(credential.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

type UserTOTPCredentialJSON struct {
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
	EncodedKey    string `json:"key"`
}
