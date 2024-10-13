package main

import (
	"bufio"
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

func handleCreateUserRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyCredential(r) {
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
		Context  *string `json:"context"`
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
	email, password, _ := *data.Email, *data.Password, data.Context
	if !verifyEmailInput(email) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidEmail)
		return
	}
	emailAvailable, err := checkEmailAvailability(email)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !emailAvailable {
		writeExpectedErrorResponse(w, ExpectedErrorEmailAlreadyUsed)
		return
	}

	if len(password) > 255 {
		writeExpectedErrorResponse(w, ExpectedErrorPasswordTooLarge)
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

	if clientIP != "" && !passwordHashingIPRateLimit.Consume(clientIP, 1) {
		logMessageWithClientIP("INFO", "CREATE_USER", "PASSWORD_HASHING_LIMIT_REJECTED", clientIP, fmt.Sprintf("email_input=\"%s\"", strings.ReplaceAll(email, "\"", "\\\"")))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	user, err := createUser(email, passwordHash)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			writeExpectedErrorResponse(w, ExpectedErrorEmailAlreadyUsed)
			return
		}
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "CREATE_USER", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s email=\"%s\"", user.Id, strings.ReplaceAll(user.Email, "\"", "\\\"")))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}

func handleGetUserRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	user, err := getUser(userId)
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

func handleDeleteUserRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	userId := params.ByName("user_id")
	err := deleteUser(userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleUpdatePasswordRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	user, err := getUser(userId)
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
	if len(password) > 255 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if len(newPassword) > 255 {
		writeExpectedErrorResponse(w, ExpectedErrorPasswordTooLarge)
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
	updateUserPassword(userId, newPasswordHash)

	w.WriteHeader(204)
}

func handleResetUser2FARequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyCredential(r) {
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
	userExists, err := checkUserExists(userId)
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
	if data.RecoveryCode == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if !recoveryCodeUserRateLimit.Consume(userId, 1) {
		logMessageWithClientIP("INFO", "RESET_2FA", "RECOVERY_CODE_USER_LIMIT_REJECTED", clientIP, fmt.Sprintf("user_id=%s", userId))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	newRecoveryCode, valid, err := resetUser2FAWithRecoveryCode(userId, *data.RecoveryCode)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !valid {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	recoveryCodeUserRateLimit.Reset(userId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	encodeRecoveryCodeToJSON(newRecoveryCode)
}

func handleRegenerateUserRecoveryCodeRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	userExists, err := checkUserExists(userId)
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	newRecoveryCode, err := regenerateUserRecoveryCode(userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	encodeRecoveryCodeToJSON(newRecoveryCode)
}

func handleGetUsersRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyCredential(r) {
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
	users, err := getUsers(sortBy, sortOrder, count, page)
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

func handleGetUserRecoveryCodeRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	recoveryCode, err := getUserRecoveryCode(userId)
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
	w.Write([]byte(encodeRecoveryCodeToJSON(recoveryCode)))
}

func createUser(email string, passwordHash string) (User, error) {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return User{}, nil
	}
	recoveryCode, err := generateSecureCode()
	if err != nil {
		return User{}, nil
	}
	_, err = db.Exec("INSERT INTO user (id, created_at, email, password_hash, recovery_code) VALUES (?, ?, ?, ?, ?)", id, now.Unix(), email, passwordHash, recoveryCode)
	if err != nil {
		return User{}, err
	}
	user := User{
		Id:           id,
		CreatedAt:    now,
		Email:        email,
		PasswordHash: passwordHash,
		RecoveryCode: recoveryCode,
	}
	return user, nil
}

func checkEmailAvailability(email string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT count(*) FROM user WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count < 1, nil
}

func getUser(userId string) (User, error) {
	var user User
	var createdAtUnix int64
	var emailVerified, registeredTOTP int
	row := db.QueryRow("SELECT user.id, user.created_at, user.email, user.password_hash, user.recovery_code, user.email_verified, IIF(totp_credential.id IS NOT NULL, 1, 0) FROM user LEFT JOIN totp_credential ON user.id = totp_credential.user_id WHERE user.id = ?", userId)
	err := row.Scan(&user.Id, &createdAtUnix, &user.Email, &user.PasswordHash, &user.RecoveryCode, &emailVerified, &registeredTOTP)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrRecordNotFound
	}
	if err != nil {
		return User{}, err
	}
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	user.EmailVerified = emailVerified == 1
	user.RegisteredTOTP = registeredTOTP == 1
	return user, nil
}

func getUserRecoveryCode(userId string) (string, error) {
	var recoveryCode string
	row := db.QueryRow("SELECT recovery_code FROM user WHERE id = ?", userId)
	err := row.Scan(&recoveryCode)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrRecordNotFound
	}
	if err != nil {
		return "", err
	}
	return recoveryCode, nil
}

func getUsers(sortBy UserSortBy, sortOrder SortOrder, count, page int) ([]User, error) {
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

	query := fmt.Sprintf(`SELECT user.id, user.created_at, user.email, user.password_hash, user.recovery_code, user.email_verified, IIF(totp_credential.id IS NOT NULL, 1, 0)
		FROM user LEFT JOIN totp_credential ON user.id = totp_credential.user_id
		ORDER BY %s %s LIMIT ? OFFSET ?`, orderBySQL, orderSQL)

	var users []User
	rows, err := db.Query(query, count, count*(page-1))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		var createdAtUnix int64
		var emailVerified, registeredTOTP int
		err = rows.Scan(&user.Id, &createdAtUnix, &user.Email, &user.PasswordHash, &user.RecoveryCode, &emailVerified, &registeredTOTP)
		if err != nil {
			return nil, err
		}
		user.CreatedAt = time.Unix(createdAtUnix, 0)
		user.EmailVerified = emailVerified == 1
		user.RegisteredTOTP = registeredTOTP == 1
		users = append(users, user)
	}
	return users, nil
}

func checkUserExists(userId string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT count(*) FROM user WHERE id = ?", userId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func getUserFromEmail(email string) (User, error) {
	var user User
	var createdAtUnix int64
	var emailVerifiedInt, registeredTOTPInt int
	row := db.QueryRow("SELECT user.id, user.created_at, user.email, user.password_hash, user.recovery_code, user.email_verified, IIF(totp_credential.id IS NOT NULL, 1, 0) FROM user LEFT JOIN totp_credential ON user.id = totp_credential.user_id WHERE user.email = ?", email)
	err := row.Scan(&user.Id, &createdAtUnix, &user.Email, &user.PasswordHash, &user.RecoveryCode, &emailVerifiedInt, &registeredTOTPInt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrRecordNotFound
	}
	if err != nil {
		return User{}, err
	}
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	user.EmailVerified = emailVerifiedInt == 1
	user.RegisteredTOTP = registeredTOTPInt == 1
	return user, nil
}

func deleteUser(userId string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM totp_credential WHERE user_id = ?", userId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM email_verification_request WHERE user_id = ?", userId)
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

func updateUserPassword(userId string, passwordHash string) error {
	_, err := db.Exec("UPDATE user SET password_hash = ? WHERE id = ?", passwordHash, userId)
	return err
}

func resetUser2FAWithRecoveryCode(userId string, recoveryCode string) (string, bool, error) {
	newRecoveryCode, err := generateSecureCode()
	if err != nil {
		return "", false, err
	}
	tx, err := db.Begin()
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec("UPDATE user SET recovery_code = ? WHERE id = ? AND recovery_code = ?", newRecoveryCode, userId, newRecoveryCode)
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
	_, err = tx.Exec("DELETE FROM totp_credential WHERE user_id = ?", userId)
	if err != nil {
		return "", false, err
	}

	// TODO: Should existing password reset requests be invalidated?

	err = tx.Commit()
	if err != nil {
		return "", false, err
	}
	return newRecoveryCode, true, nil
}

func regenerateUserRecoveryCode(userId string) (string, error) {
	newRecoveryCode, err := generateSecureCode()
	if err != nil {
		return "", err
	}
	_, err = db.Exec("UPDATE user SET recovery_code = ? WHERE id = ?", newRecoveryCode, userId)
	if err != nil {
		return "", err
	}
	return newRecoveryCode, nil
}

var emailInputRegex = regexp.MustCompile(`^.+@.+\..+$`)

func verifyEmailInput(email string) bool {
	return len(email) < 256 && emailInputRegex.MatchString(email)
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
	EmailVerified  bool
	RegisteredTOTP bool
}

func (u *User) Registered2FA() bool {
	return u.RegisteredTOTP
}

func (u *User) EncodeToJSON() string {
	escapedEmail := strings.ReplaceAll(u.Email, "\"", "\\\"")
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"created_at\":%d,\"email\":\"%s\",\"email_verified\":%t,\"registered_totp\":%t}", u.Id, u.CreatedAt.Unix(), escapedEmail, u.EmailVerified, u.RegisteredTOTP)
	return encoded
}

type UserSortBy int

const (
	UserSortByCreatedAt UserSortBy = iota
	UserSortByEmail
	UserSortById
)
