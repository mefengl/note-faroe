package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserEmailVerificationRequest(t *testing.T) {
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

	request := UserEmailVerificationRequest{
		UserId:    user.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &request)
	if err != nil {
		t.Fatal(err)
	}

	result, err := getUserEmailVerificationRequest(db, context.Background(), request.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, request, result)

	_, err = getUser(db, context.Background(), "2")
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestCreateUserEmailVerificationRequest(t *testing.T) {
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

	request1, err := createUserEmailVerificationRequest(db, context.Background(), user.Id)
	if err != nil {
		t.Fatal(err)
	}
	result, err := getUserEmailVerificationRequest(db, context.Background(), request1.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, request1, result)

	request2, err := createUserEmailVerificationRequest(db, context.Background(), user.Id)
	if err != nil {
		t.Fatal(err)
	}
	result, err = getUserEmailVerificationRequest(db, context.Background(), request2.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, request2, result)
	assert.NotEqual(t, request1, result)
}

func TestDeleteUserEmailVerificationRequest(t *testing.T) {
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

	request1 := UserEmailVerificationRequest{
		UserId:    user1.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &request1)
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

	request2 := UserEmailVerificationRequest{
		UserId:    user2.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &request2)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUserEmailVerificationRequest(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getUserEmailVerificationRequest(db, context.Background(), request1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUserEmailVerificationRequest(db, context.Background(), request2.UserId)
	assert.Nil(t, err)
}

func insertUserEmailVerificationRequest(db *sql.DB, request *UserEmailVerificationRequest) error {
	_, err := db.Exec("INSERT INTO user_email_verification_request (user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?)", request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code, request.CreatedAt.Unix(), request.Code, request.UserId)
	return err
}

func TestGetEmailUpdateRequest(t *testing.T) {
	t.Parallel()
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user := User{
		Id:             "1",
		Email:          "user1a@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user)
	if err != nil {
		t.Fatal(err)
	}

	request := EmailUpdateRequest{
		Id:        "1",
		UserId:    user.Id,
		Email:     "user1b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &request)
	if err != nil {
		t.Fatal(err)
	}

	result, err := getEmailUpdateRequest(db, context.Background(), request.UserId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, request, result)

	_, err = getUser(db, context.Background(), "2")
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestDeleteEmailUpdateRequest(t *testing.T) {
	t.Parallel()
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user1 := User{
		Id:             "1",
		Email:          "user1a@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user1)
	if err != nil {
		t.Fatal(err)
	}

	request1 := EmailUpdateRequest{
		Id:        "1",
		UserId:    user1.Id,
		Email:     "user1b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &request1)
	if err != nil {
		t.Fatal(err)
	}

	user2 := User{
		Id:             "2",
		Email:          "user2a@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	request2 := EmailUpdateRequest{
		Id:        "2",
		UserId:    user2.Id,
		Email:     "user2b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &request2)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteEmailUpdateRequest(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getEmailUpdateRequest(db, context.Background(), request1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getEmailUpdateRequest(db, context.Background(), request2.UserId)
	assert.Nil(t, err)
}

func TestUpdateUserEmailWithUpdateRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:             "1",
			Email:          "user1a@example.com",
			CreatedAt:      now,
			PasswordHash:   "HASH1",
			RecoveryCode:   "CODE1",
			TOTPRegistered: false,
		}
		err := insertUser(db, context.Background(), &user1)
		if err != nil {
			t.Fatal(err)
		}

		request1 := EmailUpdateRequest{
			Id:        "1",
			UserId:    user1.Id,
			Email:     "user1b@example.com",
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertEmailUpdateRequest(db, context.Background(), &request1)
		if err != nil {
			t.Fatal(err)
		}

		request2 := EmailUpdateRequest{
			Id:        "2",
			UserId:    user1.Id,
			Email:     "user1c@example.com",
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertEmailUpdateRequest(db, context.Background(), &request2)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user1.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH1",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
		if err != nil {
			t.Fatal(err)
		}

		user2 := User{
			Id:             "2",
			Email:          "user2a@example.com",
			CreatedAt:      now,
			PasswordHash:   "HASH1",
			RecoveryCode:   "CODE1",
			TOTPRegistered: false,
		}
		err = insertUser(db, context.Background(), &user2)
		if err != nil {
			t.Fatal(err)
		}
		request3 := EmailUpdateRequest{
			Id:        "3",
			UserId:    user2.Id,
			Email:     "user2a@example.com",
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertEmailUpdateRequest(db, context.Background(), &request3)
		if err != nil {
			t.Fatal(err)
		}
		request4 := EmailUpdateRequest{
			Id:        "4",
			UserId:    user2.Id,
			Email:     "user1b@example.com",
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertEmailUpdateRequest(db, context.Background(), &request4)
		if err != nil {
			t.Fatal(err)
		}

		resetRequest2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user2.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH1",
		}
		err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
		if err != nil {
			t.Fatal(err)
		}

		user1.Email = request1.Email
		ok, err := updateUserEmailWithUpdateRequest(db, context.Background(), request1.Id, request1.Code)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, ok)

		_, err = getEmailUpdateRequest(db, context.Background(), request1.Id)
		assert.ErrorIs(t, err, ErrRecordNotFound)
		_, err = getEmailUpdateRequest(db, context.Background(), request2.Id)
		assert.Nil(t, err)
		_, err = getEmailUpdateRequest(db, context.Background(), request3.Id)
		assert.Nil(t, err)
		_, err = getEmailUpdateRequest(db, context.Background(), request4.Id)
		assert.ErrorIs(t, err, ErrRecordNotFound)

		_, err = getPasswordResetRequest(db, context.Background(), resetRequest1.Id)
		assert.ErrorIs(t, err, ErrRecordNotFound)
		_, err = getPasswordResetRequest(db, context.Background(), resetRequest2.Id)
		assert.Nil(t, err)

		result, err := getUser(db, context.Background(), user1.Id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user1, result)

	})
	t.Run("expired request", func(t *testing.T) {
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

		request := EmailUpdateRequest{
			Id:        "1",
			UserId:    user.Id,
			Email:     "user1b@example.com",
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(-10 * time.Minute),
		}
		err = insertEmailUpdateRequest(db, context.Background(), &request)
		if err != nil {
			t.Fatal(err)
		}

		ok, err := updateUserEmailWithUpdateRequest(db, context.Background(), request.Id, "12345678")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, ok)
	})

	t.Run("incorrect code", func(t *testing.T) {
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

		request := EmailUpdateRequest{
			Id:        "1",
			UserId:    user.Id,
			Email:     "user1b@example.com",
			CreatedAt: now,
			Code:      "12345678",
			ExpiresAt: now.Add(10 * time.Minute),
		}
		err = insertEmailUpdateRequest(db, context.Background(), &request)
		if err != nil {
			t.Fatal(err)
		}

		ok, err := updateUserEmailWithUpdateRequest(db, context.Background(), request.Id, "87654321")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, ok)
	})

	t.Run("invalid request", func(t *testing.T) {
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

		ok, err := updateUserEmailWithUpdateRequest(db, context.Background(), "1", "12345678")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, ok)
	})

}

func TestEncodeEmailToJSON(t *testing.T) {
	email := "user@example.com"

	expected := EmailJSON{
		Email: email,
	}

	var result EmailJSON

	json.Unmarshal([]byte(encodeEmailToJSON(email)), &result)

	assert.Equal(t, expected, result)
}

func TestEmailUpdateRequestEncodeToJSON(t *testing.T) {
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
