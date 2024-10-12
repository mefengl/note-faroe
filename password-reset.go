package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"faroe/otp"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleCreatePasswordResetRequestRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	var data struct {
		Email *string `json:"email"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	if data.Email == nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	email := *data.Email

	if !verifyEmailInput(email) {
		writeExpectedErrorResponse(w, expectedErrorInvalidEmail)
		return
	}

	user, err := getUserFromEmail(email)
	if errors.Is(err, ErrRecordNotFound) {
		logMessageWithClientIP("INFO", "CREATE_PASSWORD_RESET_REQUEST", "INVALID_EMAIL", clientIP, fmt.Sprintf("email_input=\"%s\"", strings.ReplaceAll(email, "\"", "\\\"")))
		writeExpectedErrorResponse(w, expectedErrorAccountNotExists)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if clientIP != "" && !passwordHashingIPRateLimit.Consume(clientIP, 1) {
		logMessageWithClientIP("INFO", "CREATE_PASSWORD_RESET_REQUEST", "PASSWORD_HASHING_LIMIT_REJECTED", clientIP, fmt.Sprintf("email_input=\"%s\"", strings.ReplaceAll(email, "\"", "\\\"")))
		writeExpectedErrorResponse(w, expectedErrorTooManyRequests)
		return
	}
	if clientIP != "" && !createPasswordResetIPRateLimit.Consume(clientIP, 1) {
		logMessageWithClientIP("INFO", "CREATE_PASSWORD_RESET_REQUEST", "CREATE_PASSWORD_RESET_REQUEST_LIMIT_REJECTED", clientIP, fmt.Sprintf("email_input=\"%s\"", strings.ReplaceAll(email, "\"", "\\\"")))
		writeExpectedErrorResponse(w, expectedErrorTooManyRequests)
		return
	}

	code, err := generateOneTimeCode()
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	codeHash, err := argon2id.Hash(code)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	err = deleteEmailUnverifiedUserPasswordResetRequests(user.Id)
	resetRequest, err := createPasswordResetRequest(user.Id, user.Email, codeHash)

	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "CREATE_PASSWORD_RESET_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s email=\"%s\" request_id=%s", user.Id, strings.ReplaceAll(user.Email, "\"", "\\\""), resetRequest.Id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(resetRequest.EncodeToJSONWithCode(code)))
}

func handleGetPasswordResetRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "GET_PASSWORD_RESET_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("request_id=%s user_id=%s", resetRequest.Id, resetRequest.UserId))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(resetRequest.EncodeToJSON()))
}

func handleGetPasswordResetRequestUserRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, user, err := getPasswordResetRequestAndUser(resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "GET_PASSWORD_RESET_USER_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("request_id=%s user_id=%s", resetRequest.Id, resetRequest.UserId))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}

func handleVerifyPasswordResetRequestEmailRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	var data struct {
		Code *string `json:"code"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	if data.Code == nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	if clientIP != "" && !passwordHashingIPRateLimit.Consume(clientIP, 1) {
		logMessageWithClientIP("INFO", "VERIFY_PASSWORD_RESET_REQUEST_EMAIL", "PASSWORD_HASHING_LIMIT_REJECTED", clientIP, "")
		writeExpectedErrorResponse(w, expectedErrorTooManyRequests)
		return
	}
	if !verifyPasswordResetCodeLimitCounter.Consume(resetRequest.Id) {
		logMessageWithClientIP("INFO", "VERIFY_PASSWORD_RESET_REQUEST_EMAIL", "FAIL_COUNTER_LIMIT_REJECTED", clientIP, "")
		err = deletePasswordResetRequest(resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, expectedErrorTooManyRequests)
		return
	}
	validCode, err := argon2id.Verify(resetRequest.CodeHash, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !validCode {
		writeExpectedErrorResponse(w, expectedErrorIncorrectCode)
		return
	}
	err = setPasswordResetRequestAsEmailVerified(resetRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func handleVerifyPasswordResetRequest2FAWithTOTPRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(resetRequestId)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
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
		Code *string `json:"code"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	if data.Code == nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	if !totpUserRateLimit.Consume(resetRequest.UserId, 1) {
		logMessageWithClientIP("INFO", "VERIFY_PASSWORD_RESET_REQUEST_2FA_TOTP", "TOTP_USER_LIMIT_REJECTED", clientIP, fmt.Sprintf("request_id=%s user_id=%s email=\"%s\"", resetRequest.Id, resetRequest.UserId, strings.ReplaceAll(resetRequest.Email, "\"", "\\\"")))
		writeExpectedErrorResponse(w, expectedErrorTooManyRequests)
		return
	}

	totpCredential, err := getUserTOTPCredential(resetRequest.Id)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, expectedErrorSecondFactorNotAllowed)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	valid := otp.VerifyTOTP(totpCredential.Key, 30*time.Second, 6, *data.Code)
	if !valid {
		writeExpectedErrorResponse(w, expectedErrorIncorrectCode)
		return
	}
	totpUserRateLimit.Reset(resetRequest.UserId)

	err = setPasswordResetRequestAsTwoFactorVerified(resetRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func handleResetPasswordRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		writeUnexpectedErrorResponse(w)
		return
	}
	var data struct {
		RequestId *string `json:"request_id"`
		Password  *string `json:"password"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}

	if data.RequestId == nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	if data.Password == nil {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	resetRequestId, password := *data.RequestId, *data.Password
	if len(password) > 255 {
		writeExpectedErrorResponse(w, expectedErrorInvalidData)
		return
	}
	strongPassword, err := verifyPasswordStrength(password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, expectedErrorWeakPassword)
		return
	}
	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	resetRequest, user, err := getPasswordResetRequestAndUser(resetRequestId)
	log.Println(err)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, expectedErrorInvalidRequestId)
		return
	}
	if err != nil {
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(resetRequestId)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, expectedErrorInvalidRequestId)
		return
	}
	if !resetRequest.EmailVerified {
		writeExpectedErrorResponse(w, expectedErrorEmailNotVerified)
		return
	}
	if user.Registered2FA() && !resetRequest.TwoFactorVerified {
		writeExpectedErrorResponse(w, expectedErrorSecondFactorNotVerified)
		return
	}

	validResetRequest, err := updateUserPasswordHashAndSetEmailAsVerifiedWithPasswordResetRequest(resetRequest.Id, passwordHash)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !validResetRequest {
		writeExpectedErrorResponse(w, expectedErrorInvalidRequestId)
		return
	}
	user, err = getUser(resetRequest.UserId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}

func createPasswordResetRequest(userId string, email string, codeHash string) (PasswordResetRequest, error) {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return PasswordResetRequest{}, nil
	}
	expiresAt := now.Add(10 * time.Minute)
	_, err = db.Exec("INSERT INTO password_reset_request (id, user_id, created_at, expires_at, email, code_hash) VALUES (?, ?, ?, ?, ?, ?)", id, userId, now.Unix(), expiresAt.Unix(), email, codeHash)
	if err != nil {
		return PasswordResetRequest{}, err
	}
	request := PasswordResetRequest{
		Id:                id,
		UserId:            userId,
		CreatedAt:         now,
		ExpiresAt:         expiresAt,
		Email:             email,
		CodeHash:          codeHash,
		EmailVerified:     false,
		TwoFactorVerified: false,
	}
	return request, nil
}

func getPasswordResetRequest(requestId string) (PasswordResetRequest, error) {
	var request PasswordResetRequest
	var createdAtUnix, expiresAtUnix int64
	var emailVerifiedInt, twoFactorVerifiedInt int
	row := db.QueryRow("SELECT id, user_id, created_at, email, code_hash, expires_at, email_verified, two_factor_verified FROM password_reset_request WHERE id = ?", requestId)
	err := row.Scan(&request.Id, &request.UserId, &createdAtUnix, &request.Email, &request.CodeHash, &expiresAtUnix, &emailVerifiedInt, &twoFactorVerifiedInt)
	if errors.Is(err, sql.ErrNoRows) {
		return PasswordResetRequest{}, ErrRecordNotFound
	}
	if err != nil {
		return PasswordResetRequest{}, err
	}
	request.CreatedAt = time.Unix(createdAtUnix, 0)
	request.ExpiresAt = time.Unix(expiresAtUnix, 0)
	request.EmailVerified = emailVerifiedInt == 1
	request.TwoFactorVerified = twoFactorVerifiedInt == 1
	return request, nil
}

func getPasswordResetRequestAndUser(requestId string) (PasswordResetRequest, User, error) {
	var request PasswordResetRequest
	var requestCreatedAtUnix, requestExpiresAtUnix int64
	var requestEmailVerifiedInt, requestTwoFactorVerifiedInt int
	var user User
	var userCreatedAtUnix int64
	var userEmailVerifiedInt, userRegisteredTOTPInt int
	row := db.QueryRow(`SELECT
	password_reset_request.id, password_reset_request.user_id, password_reset_request.created_at, password_reset_request.email, password_reset_request.code_hash, password_reset_request.expires_at, password_reset_request.email_verified, password_reset_request.two_factor_verified,
	user.id, user.created_at, user.email, user.password_hash, user.email_verified, IIF(totp_credential.id IS NOT NULL, 1, 0)
	FROM password_reset_request
	INNER JOIN user ON user.id = password_reset_request.user_id
	LEFT JOIN totp_credential ON password_reset_request.user_id = totp_credential.user_id
	WHERE password_reset_request.id = ?`, requestId)
	err := row.Scan(&request.Id, &request.UserId, &requestCreatedAtUnix, &request.Email, &request.CodeHash, &requestExpiresAtUnix, &requestEmailVerifiedInt, &requestTwoFactorVerifiedInt, &user.Id, &userCreatedAtUnix, &user.Email, &user.PasswordHash, &userEmailVerifiedInt, &userRegisteredTOTPInt)
	if errors.Is(err, sql.ErrNoRows) {
		return PasswordResetRequest{}, User{}, ErrRecordNotFound
	}
	if err != nil {
		return PasswordResetRequest{}, User{}, err
	}
	request.CreatedAt = time.Unix(requestCreatedAtUnix, 0)
	request.ExpiresAt = time.Unix(requestExpiresAtUnix, 0)
	request.EmailVerified = requestEmailVerifiedInt == 1
	request.TwoFactorVerified = requestTwoFactorVerifiedInt == 1
	user.CreatedAt = time.Unix(userCreatedAtUnix, 0)
	user.EmailVerified = userEmailVerifiedInt == 1
	user.RegisteredTOTP = userRegisteredTOTPInt == 1
	return request, user, nil
}

func updateUserPasswordHashAndSetEmailAsVerifiedWithPasswordResetRequest(requestId string, passwordHash string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	var userId string
	var expiresAtUnix int64
	var email string
	err = db.QueryRow("DELETE FROM password_reset_request WHERE id = ? RETURNING user_id, expires_at, email", requestId).Scan(&userId, &expiresAtUnix, &email)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return false, err
		}
		return false, nil
	}
	if err != nil {
		tx.Rollback()
		return false, err
	}
	if time.Now().Compare(time.Unix(expiresAtUnix, 0)) >= 0 {
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return false, err
		}
		return false, nil
	}
	_, err = db.Exec("UPDATE user SET password_hash = ? WHERE id = ?", passwordHash, userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	_, err = db.Exec("UPDATE user SET email_verified = 1 WHERE id = ? AND email = ?", passwordHash, userId, email)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	tx.Commit()
	return true, nil
}

func setPasswordResetRequestAsEmailVerified(requestId string) error {
	_, err := db.Exec("UPDATE password_reset_request SET email_verified = 1 WHERE id = ?", requestId)
	return err
}

func setPasswordResetRequestAsTwoFactorVerified(requestId string) error {
	_, err := db.Exec("UPDATE password_reset_request SET two_factor_verified = 1 WHERE id = ?", requestId)
	return err
}

func deletePasswordResetRequest(requestId string) error {
	_, err := db.Exec("DELETE FROM password_reset_request WHERE id = ?", requestId)
	return err
}

func deleteEmailUnverifiedUserPasswordResetRequests(userId string) error {
	_, err := db.Exec("DELETE FROM password_reset_request WHERE user_id = ? AND email_verified = 0", userId)
	return err
}

type PasswordResetRequest struct {
	Id                string
	UserId            string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	Email             string
	CodeHash          string
	EmailVerified     bool
	TwoFactorVerified bool
}

func (r *PasswordResetRequest) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d,\"email_verified\":%t,\"two_factor_verified\":%t}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix(), r.EmailVerified, r.TwoFactorVerified)
	return encoded
}

func (r *PasswordResetRequest) EncodeToJSONWithCode(code string) string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d,\"code\":\"%s\",\"email_verified\":%t,\"two_factor_verified\":%t}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix(), code, r.EmailVerified, r.TwoFactorVerified)
	return encoded
}
