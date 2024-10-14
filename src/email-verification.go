package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mattn/go-sqlite3"
)

func handleCreateEmailVerificationRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	var data struct {
		Email *string `json:"email"`
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
	email := *data.Email
	if !verifyEmailInput(email) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidEmail)
		return
	}

	if !createEmailVerificationUserRateLimit.Consume(userId, 1) {
		logMessageWithClientIP("INFO", "CREATE_EMAIL_VERIFICATION_REQUEST", "PASSWORD_HASHING_LIMIT_REJECTED", clientIP, fmt.Sprintf("user_id=%s", userId))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	verificationRequest, err := createEmailVerificationRequest(userId, email)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			writeExpectedErrorResponse(w, ExpectedErrorEmailAlreadyUsed)
			return
		}
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "CREATE_EMAIL_VERIFICATION_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s", userId))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

func handleVerifyUserEmailRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	var data struct {
		RequestId *string `json:"request_id"`
		Code      *string `json:"code"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.RequestId == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.Code == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	verificationRequest, err := getUserEmailVerificationRequest(userId, *data.RequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequestId)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		err = deleteEmailVerificationRequest(verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequestId)
		return
	}
	if !verifyEmailVerificationCodeLimitCounter.Consume(verificationRequest.Email) {
		logMessageWithClientIP("INFO", "VERIFY_EMAIL_VERIFICATION_REQUEST", "FAIL_COUNTER_LIMIT_REJECTED", clientIP, "")
		err = deleteEmailVerificationRequest(verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		logMessageWithClientIP("INFO", "VERIFY_EMAIL", "EMAIL_VERIFICATION_LIMIT_REJECTED", clientIP, fmt.Sprintf("user_id=%s request_id=%s", userId, verificationRequest.Id))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, verifiedEmail, err := validateEmailVerificationRequest(userId, verificationRequest.Id, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validCode {
		logMessageWithClientIP("INFO", "VERIFY_EMAIL", "INVALID_CODE", clientIP, fmt.Sprintf("user_id=%s", userId))
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	logMessageWithClientIP("INFO", "VERIFY_EMAIL", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s request_id=%s email=\"%s\"", userId, verificationRequest.Id, strings.ReplaceAll(verifiedEmail, "\"", "\\\"")))

	verifyEmailVerificationCodeLimitCounter.Delete(verificationRequest.Email)

	user, err := getUser(verificationRequest.UserId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}

func handleDeleteEmailVerificationRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	userId := params.ByName("user_id")
	verificationRequestId := params.ByName("request_id")
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

	err = deleteUserEmailVerificationRequest(userId, verificationRequestId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "DELETE_EMAIL_VERIFICATION_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s request_id=%s", userId, verificationRequestId))
	w.WriteHeader(204)
}

func handleGetEmailVerificationRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifyCredential(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequestId := params.ByName("request_id")
	verificationRequest, err := getUserEmailVerificationRequest(userId, verificationRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		err = deleteEmailVerificationRequest(verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "GET_EMAIL_VERIFICATION_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s request_id=%s", userId, verificationRequestId))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

func createEmailVerificationRequest(userId string, email string) (EmailVerificationRequest, error) {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return EmailVerificationRequest{}, nil
	}
	expiresAt := now.Add(10 * time.Minute)
	code, err := generateSecureCode()
	if err != nil {
		return EmailVerificationRequest{}, nil
	}
	_, err = db.Exec("INSERT INTO email_verification_request (id, user_id, created_at, expires_at, email, code) VALUES (?, ?, ?, ?, ?, ?)", id, now.Unix(), userId, expiresAt.Unix(), email, code)
	if err != nil {
		return EmailVerificationRequest{}, err
	}
	request := EmailVerificationRequest{
		Id:        id,
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Email:     email,
		Code:      code,
	}
	return request, nil
}

func getUserEmailVerificationRequest(userId string, requestId string) (EmailVerificationRequest, error) {
	var verificationRequest EmailVerificationRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRow("SELECT id, user_id, created_at, email, code, expires_at FROM email_verification_request WHERE id = ? AND user_id = ?", requestId, userId)
	err := row.Scan(&verificationRequest.Id, &verificationRequest.UserId, &createdAtUnix, &verificationRequest.Email, &verificationRequest.Code, &expiresAtUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return EmailVerificationRequest{}, ErrRecordNotFound
	}
	verificationRequest.CreatedAt = time.Unix(createdAtUnix, 0)
	verificationRequest.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return verificationRequest, nil
}

func deleteUserEmailVerificationRequests(userId string) error {
	_, err := db.Exec("DELETE FROM email_verification_request WHERE user_id = ?", userId)
	return err
}

func deleteUserEmailVerificationRequest(userId string, requestId string) error {
	_, err := db.Exec("DELETE FROM email_verification_request WHERE id = ? AND user_id = ?", requestId, userId)
	return err
}

func deleteEmailVerificationRequest(userId string) error {
	_, err := db.Exec("DELETE FROM email_verification_request WHERE user_id = ?", userId)
	return err
}

func validateEmailVerificationRequest(userId string, requestId string, code string) (bool, string, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, "", err
	}
	row := tx.QueryRow("DELETE FROM email_verification_request WHERE id = ? AND user_id = ? AND code = ? AND expires_at > ? RETURNING email", requestId, userId, code, time.Now().Unix())
	var email string
	err = row.Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return false, "", err
		}
		return false, "", nil
	}
	if err != nil {
		tx.Rollback()
		return false, "", err
	}
	_, err = tx.Exec("UPDATE user SET email = ?, email_verified = 1 WHERE id = ?", email, userId)
	if err != nil {
		tx.Rollback()
		return false, "", err
	}
	_, err = tx.Exec("DELETE FROM password_reset_request WHERE user_id = ?", userId)
	if err != nil {
		tx.Rollback()
		return false, "", err
	}
	_, err = tx.Exec("DELETE FROM email_verification_request WHERE email = ?", email)
	if err != nil {
		tx.Rollback()
		return false, "", err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return false, "", err
	}
	tx.Commit()
	return true, email, nil
}

type EmailVerificationRequest struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	Email     string
	Code      string
	ExpiresAt time.Time
}

func (r *EmailVerificationRequest) EncodeToJSON() string {
	escapedEmail := strings.ReplaceAll(r.Email, "\"", "\\\"")
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"email\":\"%s\",\"expires_at\":%d,\"code\":\"%s\"}", r.Id, r.UserId, r.CreatedAt.Unix(), escapedEmail, r.ExpiresAt.Unix(), r.Code)
	return encoded
}
