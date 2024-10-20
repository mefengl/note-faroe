package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPasswordResetRequest(t *testing.T) {
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

	request := PasswordResetRequest{
		Id:        "1",
		UserId:    user.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH",
	}

	err = insertPasswordResetRequest(db, context.Background(), &request)
	if err != nil {
		t.Fatal(err)
	}

	result, err := getPasswordResetRequest(db, context.Background(), request.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, request, result)

	_, err = getPasswordResetRequest(db, context.Background(), "2")
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestDeletePasswordResetRequest(t *testing.T) {
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

	request1 := PasswordResetRequest{
		Id:        "1",
		UserId:    user.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &request1)
	if err != nil {
		t.Fatal(err)
	}

	request2 := PasswordResetRequest{
		Id:        "2",
		UserId:    user.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &request2)
	if err != nil {
		t.Fatal(err)
	}

	err = deletePasswordResetRequest(db, context.Background(), request1.Id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getPasswordResetRequest(db, context.Background(), request1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getPasswordResetRequest(db, context.Background(), request2.Id)
	assert.Nil(t, err)
}

func TestResetUserPasswordWithPasswordResetRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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

		request1 := PasswordResetRequest{
			Id:        "1",
			UserId:    user1.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH1",
		}
		err = insertPasswordResetRequest(db, context.Background(), &request1)
		if err != nil {
			t.Fatal(err)
		}

		request2 := PasswordResetRequest{
			Id:        "2",
			UserId:    user1.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH1",
		}
		err = insertPasswordResetRequest(db, context.Background(), &request2)
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
		request3 := PasswordResetRequest{
			Id:        "3",
			UserId:    user2.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
			CodeHash:  "HASH1",
		}
		err = insertPasswordResetRequest(db, context.Background(), &request3)
		if err != nil {
			t.Fatal(err)
		}

		user1.PasswordHash += "+"
		ok, err := resetUserPasswordWithPasswordResetRequest(db, context.Background(), request1.Id, user1.PasswordHash)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, ok)

		_, err = getPasswordResetRequest(db, context.Background(), request1.Id)
		assert.ErrorIs(t, err, ErrRecordNotFound)
		_, err = getPasswordResetRequest(db, context.Background(), request2.Id)
		assert.ErrorIs(t, err, ErrRecordNotFound)
		_, err = getPasswordResetRequest(db, context.Background(), request3.Id)
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

		request := PasswordResetRequest{
			Id:        "1",
			UserId:    user.Id,
			CreatedAt: now,
			ExpiresAt: now.Add(-10 * time.Minute),
			CodeHash:  "HASH1",
		}
		err = insertPasswordResetRequest(db, context.Background(), &request)
		if err != nil {
			t.Fatal(err)
		}

		ok, err := resetUserPasswordWithPasswordResetRequest(db, context.Background(), request.Id, "HASH1+")
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

		ok, err := resetUserPasswordWithPasswordResetRequest(db, context.Background(), "1", "HASH1+")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, ok)
	})

}

func TestPasswordResetRequestEncodeToJSON(t *testing.T) {
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

type PasswordResetRequestJSON struct {
	Id            string `json:"id"`
	UserId        string `json:"user_id"`
	CreatedAtUnix int64  `json:"created_at"`
	ExpiresAtUnix int64  `json:"expires_at"`
}
