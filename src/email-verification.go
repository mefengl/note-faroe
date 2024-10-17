package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleCreateUserEmailVerificationRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifySecret(r) {
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

	if !createEmailVerificationUserRateLimit.Consume(userId, 1) {
		logMessageWithClientIP("INFO", "CREATE_EMAIL_VERIFICATION_REQUEST", "PASSWORD_HASHING_LIMIT_REJECTED", clientIP, fmt.Sprintf("user_id=%s", userId))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	verificationRequest, err := createEmailVerificationRequest(userId)
	if err != nil {
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

	if !verifySecret(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequest, err := getUserEmailVerificationRequest(userId)
	if errors.Is(err, ErrRecordNotFound) {
		createEmailVerificationUserRateLimit.AddToken(userId, 1)
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
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
		createEmailVerificationUserRateLimit.AddToken(userId, 1)
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	var data struct {
		Code *string `json:"code"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.Code == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if !verifyUserEmailRateLimit.Consume(userId, 1) {
		logMessageWithClientIP("INFO", "VERIFY_EMAIL_VERIFICATION_REQUEST", "FAIL_COUNTER_LIMIT_REJECTED", clientIP, "")
		err = deleteEmailVerificationRequest(verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		logMessageWithClientIP("INFO", "VERIFY_EMAIL", "EMAIL_VERIFICATION_LIMIT_REJECTED", clientIP, fmt.Sprintf("user_id=%s request_id=%s", verificationRequest.UserId, verificationRequest.Id))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := validateUserEmailVerificationRequest(userId, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validCode {
		logMessageWithClientIP("INFO", "VERIFY_EMAIL", "INVALID_CODE", clientIP, fmt.Sprintf("user_id=%s", verificationRequest.UserId))
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	verifyUserEmailRateLimit.Reset(verificationRequest.UserId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}

func handleDeleteUserEmailVerificationRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifySecret(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequest, err := getUserEmailVerificationRequest(userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	err = deleteEmailVerificationRequest(verificationRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	logMessageWithClientIP("INFO", "DELETE_EMAIL_VERIFICATION_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s request_id=%s", verificationRequest.UserId, verificationRequest.Id))
	w.WriteHeader(204)
}

func handleGetUserEmailVerificationRequestRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")

	if !verifySecret(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequest, err := getUserEmailVerificationRequest(userId)
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
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	logMessageWithClientIP("INFO", "GET_EMAIL_VERIFICATION_REQUEST", "SUCCESS", clientIP, fmt.Sprintf("user_id=%s request_id=%s", verificationRequest.UserId, verificationRequest.Id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

func createEmailVerificationRequest(userId string) (EmailVerificationRequest, error) {
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
	_, err = db.Exec(`INSERT INTO email_verification_request (id, user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET id = ?, created_at = ?, code = ? WHERE user_id = ?`, id, userId, now.Unix(), expiresAt.Unix(), code, id, now.Unix(), code, userId)
	if err != nil {
		return EmailVerificationRequest{}, err
	}
	request := EmailVerificationRequest{
		Id:        id,
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Code:      code,
	}
	return request, nil
}

func getUserEmailVerificationRequest(userId string) (EmailVerificationRequest, error) {
	var verificationRequest EmailVerificationRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRow("SELECT id, user_id, created_at, expires_at, code FROM email_verification_request WHERE user_id = ?", userId)
	err := row.Scan(&verificationRequest.Id, &verificationRequest.UserId, &createdAtUnix, &expiresAtUnix, &verificationRequest.Code)
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

func deleteEmailVerificationRequest(requestId string) error {
	_, err := db.Exec("DELETE FROM email_verification_request WHERE id = ?", requestId)
	return err
}

func validateUserEmailVerificationRequest(userId string, code string) (bool, error) {
	result, err := db.Exec("DELETE FROM email_verification_request WHERE user_id = ? AND code = ? AND expires_at > ?", userId, code, time.Now().Unix())
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

type EmailVerificationRequest struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	Code      string
	ExpiresAt time.Time
}

func (r *EmailVerificationRequest) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d,\"code\":\"%s\"}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix(), r.Code)
	return encoded
}
