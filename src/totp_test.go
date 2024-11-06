package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTOTPCredentialEncodeToJSON(t *testing.T) {
	t.Parallel()

	now := time.Unix(time.Now().Unix(), 0)

	credential := TOTPCredential{
		Id:        "1",
		UserId:    "1",
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}

	expected := TOTPCredentialJSON{
		Id:            "1",
		UserId:        credential.UserId,
		CreatedAtUnix: credential.CreatedAt.Unix(),
	}

	var result TOTPCredentialJSON

	json.Unmarshal([]byte(credential.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

type TOTPCredentialJSON struct {
	Id            string `json:"id"`
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
}
