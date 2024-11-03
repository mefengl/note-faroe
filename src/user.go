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
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleCreateUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		writeUnexpectedErrorResponse(w)
		return
	}
	var data struct {
		Password *string `json:"password"`
		ClientIP string  `json:"client_ip"`
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

	if *data.Password == "" || len(*data.Password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	strongPassword, err := verifyPasswordStrength(*data.Password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword)
		return
	}

	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	passwordHash, err := argon2id.Hash(*data.Password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	user, err := createUser(env.db, r.Context(), passwordHash)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
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
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	err = deleteUser(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleUpdateUserPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
		writeUnexpectedErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	var data struct {
		Password    *string `json:"password"`
		NewPassword *string `json:"new_password"`
		ClientIP    string  `json:"client_ip"`
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

	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	strongPassword, err := verifyPasswordStrength(newPassword)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword)
		return
	}

	validPassword, err := argon2id.Verify(user.PasswordHash, password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !validPassword {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}
	newPasswordHash, err := argon2id.Hash(newPassword)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	updateUserPassword(env.db, r.Context(), userId, newPasswordHash)

	w.WriteHeader(204)
}

func handleVerifyUserRecoveryCodeRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
	newRecoveryCode, valid, err := verifyUserRecoveryCode(env.db, r.Context(), userId, *data.RecoveryCode)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	newRecoveryCode, err := regenerateUserRecoveryCode(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(encodeRecoveryCodeToJSON(newRecoveryCode)))
}

func handleDeleteUserSecondFactorsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	err = deleteUserSecondFactors(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetUsersRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	responseContentType, ok := parseJSONOrTextAcceptHeader(r)
	if !ok {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	var sortBy UserSortBy
	sortByQuery := r.URL.Query().Get("sort_by")
	if sortByQuery == "created_at" {
		sortBy = UserSortByCreatedAt
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

	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage < 1 {
		perPage = 20
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	users, userCount, err := getUsers(env.db, r.Context(), sortBy, sortOrder, perPage, page)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	totalPages := int64(math.Ceil(float64(userCount) / float64(perPage)))
	if responseContentType == ContentTypeJSON {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Pagination-Total", strconv.FormatInt(userCount, 10))
		w.Header().Set("X-Pagination-Total-Pages", strconv.FormatInt(totalPages, 10))
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
		return
	}
	if responseContentType == ContentTypePlainText {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Pagination-Total", strconv.FormatInt(userCount, 10))
		w.Header().Set("X-Pagination-Total-Pages", strconv.FormatInt(totalPages, 10))
		w.WriteHeader(200)
		writeUserListAsFormattedString(w, users)
		return
	}
	writeUnexpectedErrorResponse(w)
}

func handleDeleteUsersRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	err := deleteUsers(env.db, r.Context())
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func createUser(db *sql.DB, ctx context.Context, passwordHash string) (User, error) {
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
	_, err := db.ExecContext(ctx, "INSERT INTO user (id, created_at, password_hash, recovery_code) VALUES (?, ?, ?, ?)", user.Id, user.CreatedAt.Unix(), user.PasswordHash, user.RecoveryCode)
	return err
}

func getUser(db *sql.DB, ctx context.Context, userId string) (User, error) {
	var user User
	var createdAtUnix int64
	var totpRegisteredInt int
	row := db.QueryRowContext(ctx, "SELECT user.id, user.created_at, user.password_hash, user.recovery_code, IIF(totp_credential.user_id IS NOT NULL, 1, 0) FROM user LEFT JOIN totp_credential ON user.id = totp_credential.user_id WHERE user.id = ?", userId)
	err := row.Scan(&user.Id, &createdAtUnix, &user.PasswordHash, &user.RecoveryCode, &totpRegisteredInt)
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

func getUsers(db *sql.DB, ctx context.Context, sortBy UserSortBy, sortOrder SortOrder, perPage, page int) ([]User, int64, error) {
	var orderBySQL, orderSQL string

	if sortBy == UserSortByCreatedAt {
		orderBySQL = "user.created_at"
	} else if sortBy == UserSortById {
		orderBySQL = "user.id"
	} else {
		return nil, 0, errors.New("invalid 'sortBy' value")
	}

	if sortOrder == SortOrderAscending {
		orderSQL = "ASC"
	} else if sortOrder == SortOrderDescending {
		orderSQL = "DESC"
	} else {
		return nil, 0, errors.New("invalid 'sortOrder' value")
	}

	query := fmt.Sprintf(`SELECT user.id, user.created_at, user.password_hash, user.recovery_code, IIF(totp_credential.user_id IS NOT NULL, 1, 0)
		FROM user LEFT JOIN totp_credential ON user.id = totp_credential.user_id
		ORDER BY %s %s LIMIT ? OFFSET ?`, orderBySQL, orderSQL)

	var users []User
	rows, err := db.QueryContext(ctx, query, perPage, perPage*(page-1))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		var createdAtUnix int64
		var totpRegisteredInt int
		err = rows.Scan(&user.Id, &createdAtUnix, &user.PasswordHash, &user.RecoveryCode, &totpRegisteredInt)
		if err != nil {
			return nil, 0, err
		}
		user.CreatedAt = time.Unix(createdAtUnix, 0)
		user.TOTPRegistered = totpRegisteredInt == 1
		users = append(users, user)
	}

	var total int64
	err = db.QueryRowContext(ctx, "SELECT count(*) FROM user").Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func deleteUsers(db *sql.DB, ctx context.Context) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM totp_credential")
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

func deleteUserSecondFactors(db *sql.DB, ctx context.Context, userId string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM totp_credential WHERE user_id = ?", userId)
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

func deleteUser(db *sql.DB, ctx context.Context, userId string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM totp_credential WHERE user_id = ?", userId)
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

func verifyUserRecoveryCode(db *sql.DB, ctx context.Context, userId string, recoveryCode string) (string, bool, error) {
	newRecoveryCode, err := generateSecureCode()
	if err != nil {
		return "", false, err
	}
	result, err := db.Exec("UPDATE user SET recovery_code = ? WHERE id = ? AND recovery_code = ?", newRecoveryCode, userId, recoveryCode)
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
	PasswordHash   string
	RecoveryCode   string
	TOTPRegistered bool
}

func (u *User) Registered2FA() bool {
	return u.TOTPRegistered
}

func (u *User) EncodeToJSON() string {
	encoded := fmt.Sprintf(`{"id":"%s","created_at":%d,"recovery_code":"%s","totp_registered":%t}`, u.Id, u.CreatedAt.Unix(), u.RecoveryCode, u.TOTPRegistered)
	return encoded
}

func writeUserListAsFormattedString(w io.Writer, users []User) {
	var timeLayout = "Jan 02 2006 15:04:05"

	w.Write([]byte(padEnd("User ID", 24)))
	w.Write(([]byte("  ")))
	w.Write([]byte(padEnd("Created at", len(timeLayout))))
	w.Write(([]byte("  ")))
	w.Write([]byte("Recovery code"))
	w.Write(([]byte("  ")))
	w.Write([]byte("TOTP"))
	w.Write([]byte("\n"))

	for _, user := range users {
		w.Write([]byte(user.Id))
		w.Write(([]byte("  ")))
		w.Write([]byte(user.CreatedAt.UTC().Format(timeLayout)))
		w.Write(([]byte("  ")))
		w.Write([]byte(padEnd(user.RecoveryCode, len("Recovery code"))))
		w.Write(([]byte("  ")))
		if user.TOTPRegistered {
			w.Write([]byte("✓"))
		} else {
			w.Write([]byte("-"))
		}
		w.Write([]byte("\n"))
	}
}

type UserSortBy int

const (
	UserSortByCreatedAt UserSortBy = iota
	UserSortById
)
