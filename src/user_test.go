package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	t.Parallel()

	t.Run("totp not registered", func(t *testing.T) {
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

		result, err := getUser(db, context.Background(), user.Id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user, result)

		_, err = getUser(db, context.Background(), "2")
		assert.ErrorIs(t, err, ErrRecordNotFound)
	})

	t.Run("totp registered", func(t *testing.T) {
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

		result, err := getUser(db, context.Background(), user.Id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user, result)

		_, err = getUser(db, context.Background(), "2")
		assert.ErrorIs(t, err, ErrRecordNotFound)
	})
}

func TestCheckEmailAvailability(t *testing.T) {
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

	result, err := checkEmailAvailability(db, context.Background(), user.Email)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, result)

	result, err = checkEmailAvailability(db, context.Background(), "user2@example.com")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, result)
}

func TestGetUserRecoveryCode(t *testing.T) {
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

	result, err := getUserRecoveryCode(db, context.Background(), user.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user.RecoveryCode, result)

	_, err = getUserRecoveryCode(db, context.Background(), "2")
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestGetUsers(t *testing.T) {
	t.Parallel()
	db := initializeTestDB(t)
	defer db.Close()

	now := time.Unix(time.Now().Unix(), 0)

	user2 := User{
		Id:             "2",
		Email:          "c@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH2",
		RecoveryCode:   "CODE2",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	user1 := User{
		Id:             "1",
		Email:          "a@example.com",
		CreatedAt:      time.Unix(now.Add(1*time.Second).Unix(), 0),
		PasswordHash:   "HASH1",
		RecoveryCode:   "CODE1",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user1)
	if err != nil {
		t.Fatal(err)
	}

	user3 := User{
		Id:           "3",
		Email:        "b@example.com",
		CreatedAt:    time.Unix(now.Add(2*time.Second).Unix(), 0),
		PasswordHash: "HASH3",
		RecoveryCode: "CODE3",
	}
	err = insertUser(db, context.Background(), &user3)
	if err != nil {
		t.Fatal(err)
	}

	result, err := getUsers(db, context.Background(), UserSortByCreatedAt, SortOrderAscending, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user2, user1, user3}, result)
	result, err = getUsers(db, context.Background(), UserSortByCreatedAt, SortOrderDescending, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user3, user1, user2}, result)

	result, err = getUsers(db, context.Background(), UserSortByEmail, SortOrderAscending, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user1, user3, user2}, result)
	result, err = getUsers(db, context.Background(), UserSortByEmail, SortOrderDescending, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user2, user3, user1}, result)

	result, err = getUsers(db, context.Background(), UserSortById, SortOrderAscending, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user1, user2, user3}, result)
	result, err = getUsers(db, context.Background(), UserSortById, SortOrderDescending, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user3, user2, user1}, result)

	result, err = getUsers(db, context.Background(), UserSortById, SortOrderAscending, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user1, user2}, result)
	result, err = getUsers(db, context.Background(), UserSortById, SortOrderAscending, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []User{user3}, result)
	result, err = getUsers(db, context.Background(), UserSortById, SortOrderAscending, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, result)
}

func TestDeleteUsers(t *testing.T) {
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

	credential1 := UserTOTPCredential{
		UserId:    user1.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential1)
	if err != nil {
		t.Fatal(err)
	}

	emailVerificationRequest1 := UserEmailVerificationRequest{
		UserId:    user1.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &emailVerificationRequest1)
	if err != nil {
		t.Fatal(err)
	}

	emailUpdateRequest1 := EmailUpdateRequest{
		Id:        "1",
		UserId:    user1.Id,
		Email:     "user1b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &emailUpdateRequest1)
	if err != nil {
		t.Fatal(err)
	}

	passwordResetRequest1 := PasswordResetRequest{
		Id:        "1",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(-10 * time.Minute),
		CodeHash:  "HASH1",
	}
	err = insertPasswordResetRequest(db, context.Background(), &passwordResetRequest1)
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

	credential2 := UserTOTPCredential{
		UserId:    user2.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential2)
	if err != nil {
		t.Fatal(err)
	}

	emailVerificationRequest2 := UserEmailVerificationRequest{
		UserId:    user2.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &emailVerificationRequest2)
	if err != nil {
		t.Fatal(err)
	}

	emailUpdateRequest2 := EmailUpdateRequest{
		Id:        "2",
		UserId:    user2.Id,
		Email:     "user2b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &emailUpdateRequest2)
	if err != nil {
		t.Fatal(err)
	}

	passwordResetRequest2 := PasswordResetRequest{
		Id:        "2",
		UserId:    user2.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(-10 * time.Minute),
		CodeHash:  "HASH1",
	}
	err = insertPasswordResetRequest(db, context.Background(), &passwordResetRequest2)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUsers(db, context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_, err = getUser(db, context.Background(), user1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUser(db, context.Background(), user2.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)

	_, err = getUserTOTPCredential(db, context.Background(), credential1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUserTOTPCredential(db, context.Background(), credential2.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)

	_, err = getUserEmailVerificationRequest(db, context.Background(), emailVerificationRequest1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUserEmailVerificationRequest(db, context.Background(), emailVerificationRequest2.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)

	_, err = getEmailUpdateRequest(db, context.Background(), emailUpdateRequest1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getEmailUpdateRequest(db, context.Background(), emailUpdateRequest2.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)

	_, err = getPasswordResetRequest(db, context.Background(), passwordResetRequest1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getPasswordResetRequest(db, context.Background(), passwordResetRequest2.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestDeleteUser(t *testing.T) {
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

	credential1 := UserTOTPCredential{
		UserId:    user1.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential1)
	if err != nil {
		t.Fatal(err)
	}

	emailVerificationRequest1 := UserEmailVerificationRequest{
		UserId:    user1.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &emailVerificationRequest1)
	if err != nil {
		t.Fatal(err)
	}

	emailUpdateRequest1 := EmailUpdateRequest{
		Id:        "1",
		UserId:    user1.Id,
		Email:     "user1b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &emailUpdateRequest1)
	if err != nil {
		t.Fatal(err)
	}

	passwordResetRequest1 := PasswordResetRequest{
		Id:        "1",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(-10 * time.Minute),
		CodeHash:  "HASH1",
	}
	err = insertPasswordResetRequest(db, context.Background(), &passwordResetRequest1)
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

	credential2 := UserTOTPCredential{
		UserId:    user2.Id,
		CreatedAt: now,
		Key:       []byte{0x01, 0x02, 0x03},
	}
	err = insertUserTOTPCredential(db, &credential2)
	if err != nil {
		t.Fatal(err)
	}

	emailVerificationRequest2 := UserEmailVerificationRequest{
		UserId:    user2.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: time.Unix(now.Add(10*time.Minute).Unix(), 0),
	}
	err = insertUserEmailVerificationRequest(db, &emailVerificationRequest2)
	if err != nil {
		t.Fatal(err)
	}

	emailUpdateRequest2 := EmailUpdateRequest{
		Id:        "2",
		UserId:    user2.Id,
		Email:     "user2b@example.com",
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute),
	}
	err = insertEmailUpdateRequest(db, context.Background(), &emailUpdateRequest2)
	if err != nil {
		t.Fatal(err)
	}

	passwordResetRequest2 := PasswordResetRequest{
		Id:        "2",
		UserId:    user2.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(-10 * time.Minute),
		CodeHash:  "HASH1",
	}
	err = insertPasswordResetRequest(db, context.Background(), &passwordResetRequest2)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteUser(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = getUser(db, context.Background(), user1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUser(db, context.Background(), user2.Id)
	assert.Nil(t, err)

	_, err = getUserTOTPCredential(db, context.Background(), credential1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUserTOTPCredential(db, context.Background(), credential2.UserId)
	assert.Nil(t, err)

	_, err = getUserEmailVerificationRequest(db, context.Background(), emailVerificationRequest1.UserId)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getUserEmailVerificationRequest(db, context.Background(), emailVerificationRequest2.UserId)
	assert.Nil(t, err)

	_, err = getEmailUpdateRequest(db, context.Background(), emailUpdateRequest1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getEmailUpdateRequest(db, context.Background(), emailUpdateRequest2.Id)
	assert.Nil(t, err)

	_, err = getPasswordResetRequest(db, context.Background(), passwordResetRequest1.Id)
	assert.ErrorIs(t, err, ErrRecordNotFound)
	_, err = getPasswordResetRequest(db, context.Background(), passwordResetRequest2.Id)
	assert.Nil(t, err)
}

func TestCheckUserExists(t *testing.T) {
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

	result, err := checkUserExists(db, context.Background(), user.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, result)

	result, err = checkUserExists(db, context.Background(), "2")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, result)
}

func TestGetUserFromEmail(t *testing.T) {
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

	result, err := getUserFromEmail(db, context.Background(), user.Email)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user, result)

	_, err = getUserFromEmail(db, context.Background(), "user2@example.com")
	assert.ErrorIs(t, err, ErrRecordNotFound)
}

func TestUpdateUserPassword(t *testing.T) {
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

	user2 := User{
		Id:             "2",
		Email:          "user2@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH2",
		RecoveryCode:   "CODE2",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	user2.PasswordHash = user2.PasswordHash + "+"
	err = updateUserPassword(db, context.Background(), user2.Id, user2.PasswordHash)
	if err != nil {
		t.Fatal(err)
	}

	result, err := getUser(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user1, result)

	result, err = getUser(db, context.Background(), user2.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user2, result)
}

func TestRegenerateUserRecoveryCode(t *testing.T) {
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

	user2 := User{
		Id:             "2",
		Email:          "user2@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH2",
		RecoveryCode:   "CODE2",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	recoveryCode, err := regenerateUserRecoveryCode(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}
	user1.RecoveryCode = recoveryCode

	result, err := getUser(db, context.Background(), user1.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user1, result)

	result, err = getUser(db, context.Background(), user2.Id)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user2, result)
}

func TestResetUser2FAWithRecoveryCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user1 := User{
			Id:           "1",
			Email:        "user1@example.com",
			CreatedAt:    now,
			PasswordHash: "HASH1",
			RecoveryCode: "12345678",
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
			Id:           "2",
			Email:        "user2@example.com",
			CreatedAt:    now,
			PasswordHash: "HASH2",
			RecoveryCode: "12345678",
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

		newRecoveryCode, ok, err := resetUser2FAWithRecoveryCode(db, context.Background(), user1.Id, user1.RecoveryCode)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, ok, "Valid recovery code")
		user1.RecoveryCode = newRecoveryCode

		result, err := getUser(db, context.Background(), user1.Id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user1, result)
		_, err = getUserTOTPCredential(db, context.Background(), credential1.UserId)
		assert.ErrorIs(t, err, ErrRecordNotFound)

		_, err = getUser(db, context.Background(), user2.Id)
		assert.Nil(t, err)
		_, err = getUserTOTPCredential(db, context.Background(), credential2.UserId)
		assert.Nil(t, err)
	})

	t.Run("invalid code", func(t *testing.T) {
		t.Parallel()
		db := initializeTestDB(t)
		defer db.Close()

		now := time.Unix(time.Now().Unix(), 0)

		user := User{
			Id:             "1",
			Email:          "user1@example.com",
			CreatedAt:      now,
			PasswordHash:   "HASH1",
			RecoveryCode:   "12345678",
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

		_, ok, err := resetUser2FAWithRecoveryCode(db, context.Background(), user.Id, "87654321")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, ok, "Invalid recovery code")
	})
}

func TestUserEncodeToJSON(t *testing.T) {
	now := time.Unix(time.Now().Unix(), 0)

	user := User{
		Id:             "1",
		Email:          "user1@example.com",
		CreatedAt:      now,
		PasswordHash:   "HASH1",
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}

	expected := UserJSON{
		Id:             user.Id,
		Email:          user.Email,
		CreatedAtUnix:  user.CreatedAt.Unix(),
		TOTPRegistered: user.TOTPRegistered,
	}

	var result UserJSON

	json.Unmarshal([]byte(user.EncodeToJSON()), &result)

	assert.Equal(t, expected, result)
}

func TestEncodeRecoveryCodeToJSON(t *testing.T) {
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
	Email          string `json:"email"`
	CreatedAtUnix  int64  `json:"created_at"`
	TOTPRegistered bool   `json:"totp_registered"`
}

type RecoveryCodeJSON struct {
	RecoveryCode string `json:"recovery_code"`
}
