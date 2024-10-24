package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
)

func handleCreateUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	var data struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if data.Email == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	email, password := strings.ToLower(*data.Email), *data.Password
	if !verifyEmailInput(email) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	emailAvailable, err := checkEmailAvailability(env.db, r.Context(), email)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !emailAvailable {
		writeExpectedErrorResponse(w, ExpectedErrorEmailAlreadyUsed)
		return
	}

	if password == "" || len(password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	strongPassword, err := verifyPasswordStrength(password)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword)
		return
	}

	if clientIP != "" && !env.passwordHashingIPRateLimit.Consume(clientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	user, err := createUser(env.db, r.Context(), email, passwordHash)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			writeExpectedErrorResponse(w, ExpectedErrorEmailAlreadyUsed)
			return
		}
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}

func handleGetUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	user, err := getUser(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}

func handleDeleteUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	userId := params.ByName("user_id")
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	err = deleteUser(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleUpdateUserPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	user, err := getUser(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	var data struct {
		Password    *string `json:"password"`
		NewPassword *string `json:"new_password"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	password, newPassword := *data.Password, *data.NewPassword
	if password == "" || len(password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if newPassword == "" || len(newPassword) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if !env.passwordHashingIPRateLimit.Consume(clientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	strongPassword, err := verifyPasswordStrength(newPassword)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword)
		return
	}

	validPassword, err := argon2id.Verify(user.PasswordHash, password)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validPassword {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}
	newPasswordHash, err := argon2id.Hash(newPassword)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	updateUserPassword(env.db, r.Context(), userId, newPasswordHash)

	w.WriteHeader(204)
}

func handleResetUser2FARequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	var data struct {
		RecoveryCode *string `json:"recovery_code"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.RecoveryCode == nil || *data.RecoveryCode == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if !env.recoveryCodeUserRateLimit.Consume(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	newRecoveryCode, valid, err := resetUser2FAWithRecoveryCode(env.db, r.Context(), userId, *data.RecoveryCode)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !valid {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	env.recoveryCodeUserRateLimit.Reset(userId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(encodeRecoveryCodeToJSON(newRecoveryCode)))
}

func handleRegenerateUserRecoveryCodeRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	user, err := getUser(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !user.Registered2FA() {
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	newRecoveryCode, err := regenerateUserRecoveryCode(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	encodeRecoveryCodeToJSON(newRecoveryCode)
}

func handleGetUsersRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	var sortBy UserSortBy
	sortByQuery := r.URL.Query().Get("sort_by")
	if sortByQuery == "created_at" {
		sortBy = UserSortByCreatedAt
	} else if sortByQuery == "email" {
		sortBy = UserSortByEmail
	} else if sortByQuery == "id" {
		sortBy = UserSortById
	} else {
		sortBy = UserSortByCreatedAt
	}

	var sortOrder SortOrder
	sortOrderQuery := r.URL.Query().Get("sort_order")
	if sortOrderQuery == "ascending" {
		sortOrder = SortOrderAscending
	} else if sortOrderQuery == "descending" {
		sortOrder = SortOrderDescending
	} else {
		sortOrder = SortOrderAscending
	}

	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil || count < 1 {
		count = 20
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	users, err := getUsers(env.db, r.Context(), sortBy, sortOrder, count, page)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if len(users) == 0 {
		w.Write([]byte("[]"))
		return
	}
	w.Write([]byte("["))
	for i, user := range users {
		w.Write([]byte(user.EncodeToJSON()))
		if i != len(users)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))
}

func handleDeleteUsersRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	err := deleteUsers(env.db, r.Context())
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func createUser(db *sql.DB, ctx context.Context, email string, passwordHash string) (User, error) {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return User{}, nil
	}
	recoveryCode, err := generateSecureCode()
	if err != nil {
		return User{}, nil
	}
	user := User{
		Id:           id,
		CreatedAt:    now,
		Email:        email,
		PasswordHash: passwordHash,
		RecoveryCode: recoveryCode,
	}
	err = insertUser(db, ctx, &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func insertUser(db *sql.DB, ctx context.Context, user *User) error {
	_, err := db.ExecContext(ctx, "INSERT INTO user (id, created_at, email, password_hash, recovery_code) VALUES (?, ?, ?, ?, ?)", user.Id, user.CreatedAt.Unix(), user.Email, user.PasswordHash, user.RecoveryCode)
	return err
}

func checkEmailAvailability(db *sql.DB, ctx context.Context, email string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT count(*) FROM user WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count < 1, nil
}

func getUser(db *sql.DB, ctx context.Context, userId string) (User, error) {
	var user User
	var createdAtUnix int64
	var totpRegisteredInt int
	row := db.QueryRowContext(ctx, "SELECT user.id, user.created_at, user.email, user.password_hash, user.recovery_code, IIF(user_totp_credential.user_id IS NOT NULL, 1, 0) FROM user LEFT JOIN user_totp_credential ON user.id = user_totp_credential.user_id WHERE user.id = ?", userId)
	err := row.Scan(&user.Id, &createdAtUnix, &user.Email, &user.PasswordHash, &user.RecoveryCode, &totpRegisteredInt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrRecordNotFound
	}
	if err != nil {
		return User{}, err
	}
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	user.TOTPRegistered = totpRegisteredInt == 1
	return user, nil
}

func getUsers(db *sql.DB, ctx context.Context, sortBy UserSortBy, sortOrder SortOrder, count, page int) ([]User, error) {
	var orderBySQL, orderSQL string

	if sortBy == UserSortByCreatedAt {
		orderBySQL = "user.created_at"
	} else if sortBy == UserSortByEmail {
		orderBySQL = "user.email"
	} else if sortBy == UserSortById {
		orderBySQL = "user.id"
	} else {
		return nil, errors.New("invalid 'sortBy' value")
	}

	if sortOrder == SortOrderAscending {
		orderSQL = "ASC"
	} else if sortOrder == SortOrderDescending {
		orderSQL = "DESC"
	} else {
		return nil, errors.New("invalid 'sortOrder' value")
	}

	query := fmt.Sprintf(`SELECT user.id, user.created_at, user.email, user.password_hash, user.recovery_code, IIF(user_totp_credential.user_id IS NOT NULL, 1, 0)
		FROM user LEFT JOIN user_totp_credential ON user.id = user_totp_credential.user_id
		ORDER BY %s %s LIMIT ? OFFSET ?`, orderBySQL, orderSQL)

	var users []User
	rows, err := db.QueryContext(ctx, query, count, count*(page-1))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		var createdAtUnix int64
		var totpRegisteredInt int
		err = rows.Scan(&user.Id, &createdAtUnix, &user.Email, &user.PasswordHash, &user.RecoveryCode, &totpRegisteredInt)
		if err != nil {
			return nil, err
		}
		user.CreatedAt = time.Unix(createdAtUnix, 0)
		user.TOTPRegistered = totpRegisteredInt == 1
		users = append(users, user)
	}
	return users, nil
}

func deleteUsers(db *sql.DB, ctx context.Context) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM user_totp_credential")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM user_email_verification_request")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM email_update_request")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM password_reset_request")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM user")
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func checkUserExists(db *sql.DB, ctx context.Context, userId string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT count(*) FROM user WHERE id = ?", userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func getUserFromEmail(db *sql.DB, ctx context.Context, email string) (User, error) {
	var user User
	var createdAtUnix int64
	var totpRegisteredInt int
	row := db.QueryRowContext(ctx, "SELECT user.id, user.created_at, user.email, user.password_hash, user.recovery_code, IIF(user_totp_credential.user_id IS NOT NULL, 1, 0) FROM user LEFT JOIN user_totp_credential ON user.id = user_totp_credential.user_id WHERE user.email = ?", email)
	err := row.Scan(&user.Id, &createdAtUnix, &user.Email, &user.PasswordHash, &user.RecoveryCode, &totpRegisteredInt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrRecordNotFound
	}
	if err != nil {
		return User{}, err
	}
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	user.TOTPRegistered = totpRegisteredInt == 1
	return user, nil
}

func deleteUser(db *sql.DB, ctx context.Context, userId string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM user_totp_credential WHERE user_id = ?", userId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM user_email_verification_request WHERE user_id = ?", userId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM email_update_request WHERE user_id = ?", userId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM password_reset_request WHERE user_id = ?", userId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM user WHERE id = ?", userId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func updateUserPassword(db *sql.DB, ctx context.Context, userId string, passwordHash string) error {
	_, err := db.ExecContext(ctx, "UPDATE user SET password_hash = ? WHERE id = ?", passwordHash, userId)
	return err
}

func resetUser2FAWithRecoveryCode(db *sql.DB, ctx context.Context, userId string, recoveryCode string) (string, bool, error) {
	newRecoveryCode, err := generateSecureCode()
	if err != nil {
		return "", false, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec("UPDATE user SET recovery_code = ? WHERE id = ? AND recovery_code = ?", newRecoveryCode, userId, recoveryCode)
	if err != nil {
		return "", false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return "", false, err
	}
	if affected < 1 {
		return "", false, err
	}
	_, err = tx.Exec("DELETE FROM user_totp_credential WHERE user_id = ?", userId)
	if err != nil {
		return "", false, err
	}

	err = tx.Commit()
	if err != nil {
		return "", false, err
	}
	return newRecoveryCode, true, nil
}

func regenerateUserRecoveryCode(db *sql.DB, ctx context.Context, userId string) (string, error) {
	newRecoveryCode, err := generateSecureCode()
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, "UPDATE user SET recovery_code = ? WHERE id = ?", newRecoveryCode, userId)
	if err != nil {
		return "", err
	}
	return newRecoveryCode, nil
}

var emailInputRegex = regexp.MustCompile(`^.+@.+\..+$`)

func verifyEmailInput(email string) bool {
	return len(email) < 256 && emailInputRegex.MatchString(email) && strings.TrimSpace(email) == email
}

func verifyPasswordStrength(password string) (bool, error) {
	if len(password) < 8 {
		return false, nil
	}
	passwordHashBytes := sha1.Sum([]byte(password))
	passwordHash := hex.EncodeToString(passwordHashBytes[:])
	hashPrefix := passwordHash[0:5]
	res, err := http.DefaultClient.Get(fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", hashPrefix))
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		hashSuffix := strings.ToLower(scanner.Text()[:35])
		if passwordHash == hashPrefix+hashSuffix {
			return false, nil
		}
	}
	return true, nil
}

func encodeRecoveryCodeToJSON(code string) string {
	encoded := fmt.Sprintf("{\"recovery_code\":\"%s\"}", code)
	return encoded
}

type User struct {
	Id             string
	CreatedAt      time.Time
	Email          string
	PasswordHash   string
	RecoveryCode   string
	TOTPRegistered bool
}

func (u *User) Registered2FA() bool {
	return u.TOTPRegistered
}

func (u *User) EncodeToJSON() string {
	escapedEmail := strings.ReplaceAll(u.Email, "\"", "\\\"")
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"created_at\":%d,\"email\":\"%s\",\"recovery_code\":\"%s\",\"totp_registered\":%t}", u.Id, u.CreatedAt.Unix(), escapedEmail, u.RecoveryCode, u.TOTPRegistered)
	return encoded
}

type UserSortBy int

const (
	UserSortByCreatedAt UserSortBy = iota
	UserSortByEmail
	UserSortById
)
