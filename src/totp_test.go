package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserTOTPCredential(t *testing.T) {
	t.Parallel()
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user := User{
		Id:             "1",
		Email:          "user1@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user)
	if err != nil {
		t.Fatal(err)
	}

	credential := UserTOTPCredential{
		UserId:    user.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential)
	if err != nil {
		t.Fatal(err)
	}

	result, err := getUserTOTPCredential(db, context.Background(), credential.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, credential, result)

	_, err = getUser(db, context.Background(), "2")
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestRegisterUserTOTPCredential(t *testing.T) {
	t.Parallel()
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user := User{
		Id:             "1",
		Email:          "user1@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user)
	if err != nil {
		t.Fatal(err)
	}

	credential1, err := registerUserTOTPCredential(db, context.Background(), user.Id, []byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatal(err)
	}
	result, err := getUserTOTPCredential(db, context.Background(), credential1.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, credential1, result)

	credential2, err := registerUserTOTPCredential(db, context.Background(), user.Id, []byte{0x04, 0x05, 0x06})
	if err != nil {
		t.Fatal(err)
	}
	result, err = getUserTOTPCredential(db, context.Background(), credential1.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, credential2, result)
	assert.NotEqual(t, credential1, result)
}

func TestDeleteUserTOTPCredential(t *testing.T) {
	t.Parallel()
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user1 := User{
		Id:             "1",
		Email:          "user1@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user1)
	if err != nil {
		t.Fatal(err)
	}

	credential1 := UserTOTPCredential{
		UserId:    user1.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential1)
	if err != nil {
		t.Fatal(err)
	}

	user2 := User{
		Id:             "2",
		Email:          "user2@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	credential2 := UserTOTPCredential{
		UserId:    user2.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential2)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserTOTPCredential(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getUserTOTPCredential(db, context.Background(), credential1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUserTOTPCredential(db, context.Background(), credential2.UserId)
	assert.Nil(t, err)
}

func insertUserTOTPCredential(db *sql.DB, credential *UserTOTPCredential) error {
	_, err := db.Exec("INSERT INTO user_totp_credential (user_id, created_at, key) VALUES (?, ?, ?)", credential.UserId, credential.CreatedAt.Unix(), credential.Key)
	return err
}

func TestUserTOTPCredentialEncodeToJSON(t *testing.T) {
	now := time.Unix(time.Now().Unix(), 0)

	credential := UserTOTPCredential{
		UserId:    "1",
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}

	expected := TOTPCredentialJSON{
		UserId:        credential.UserId,
		CreatedAtUnix: credential.CreatedAt.Unix(),
		EncodedKey:    base64.StdEncoding.EncodeToString(credential.Key),
	}

	var result TOTPCredentialJSON

	json.Unmarshal([]byte(credential.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

type TOTPCredentialJSON struct {
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
	EncodedKey    string `json:"key"`
}
