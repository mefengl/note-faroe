package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleCreatePasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	user, err := getUserFromEmail(env.db, r.Context(), email)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorUserNotExists)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if clientIP != "" && !env.passwordHashingIPRateLimit.Consume(clientIP, 1) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if clientIP != "" && !env.createPasswordResetIPRateLimit.Consume(clientIP, 1) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	code, err := generateSecureCode()
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	codeHash, err := argon2id.Hash(code)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	resetRequest, err := createPasswordResetRequest(env.db, r.Context(), user.Id, codeHash)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(resetRequest.EncodeToJSONWithCode(code)))
}

func handleGetPasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
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
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(resetRequest.EncodeToJSON()))
}

func handleVerifyPasswordResetRequestEmailRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clientIP := r.Header.Get("X-Client-IP")
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
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
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
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
	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if clientIP != "" && !env.passwordHashingIPRateLimit.Consume(clientIP, 1) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if !env.verifyPasswordResetCodeLimitCounter.Consume(resetRequest.Id) {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := argon2id.Verify(resetRequest.CodeHash, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validCode {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	w.WriteHeader(204)
}

func handleResetPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	var data struct {
		RequestId *string `json:"request_id"`
		Password  *string `json:"password"`
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
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	resetRequestId, password := *data.RequestId, *data.Password
	if len(password) > 127 {
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
	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
	log.Println(err)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}
	if err != nil {
		writeUnExpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequestId)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}

	validResetRequest, err := resetUserPasswordWithPasswordResetRequest(env.db, r.Context(), resetRequest.Id, passwordHash)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validResetRequest {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}

	w.WriteHeader(204)
}

func handleDeletePasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
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
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func createPasswordResetRequest(db *sql.DB, ctx context.Context, userId string, codeHash string) (PasswordResetRequest, error) {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return PasswordResetRequest{}, nil
	}

	request := PasswordResetRequest{
		Id:        id,
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		CodeHash:  codeHash,
	}
	err = insertPasswordResetRequest(db, ctx, &request)
	if err != nil {
		return PasswordResetRequest{}, err
	}
	return request, nil
}

func insertPasswordResetRequest(db *sql.DB, ctx context.Context, request *PasswordResetRequest) error {
	_, err := db.ExecContext(ctx, "INSERT INTO password_reset_request (id, user_id, created_at, expires_at, code_hash) VALUES (?, ?, ?, ?, ?)", request.Id, request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.CodeHash)
	return err
}

func getPasswordResetRequest(db *sql.DB, ctx context.Context, requestId string) (PasswordResetRequest, error) {
	var request PasswordResetRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT id, user_id, created_at, code_hash, expires_at FROM password_reset_request WHERE id = ?", requestId)
	err := row.Scan(&request.Id, &request.UserId, &createdAtUnix, &request.CodeHash, &expiresAtUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return PasswordResetRequest{}, ErrRecordNotFound
	}
	if err != nil {
		return PasswordResetRequest{}, err
	}
	request.CreatedAt = time.Unix(createdAtUnix, 0)
	request.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return request, nil
}

func resetUserPasswordWithPasswordResetRequest(db *sql.DB, ctx context.Context, requestId string, passwordHash string) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	var userId string
	err = tx.QueryRow("DELETE FROM password_reset_request WHERE id = ? AND expires_at > ? RETURNING user_id", requestId, time.Now().Unix()).Scan(&userId)
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
	_, err = tx.Exec("DELETE FROM password_reset_request WHERE user_id = ?", userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	_, err = tx.Exec("UPDATE user SET password_hash = ? WHERE id = ?", passwordHash, userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	tx.Commit()
	return true, nil
}

func deletePasswordResetRequest(db *sql.DB, ctx context.Context, requestId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM password_reset_request WHERE id = ?", requestId)
	return err
}

type PasswordResetRequest struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	ExpiresAt time.Time
	CodeHash  string
}

func (r *PasswordResetRequest) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix())
	return encoded
}

func (r *PasswordResetRequest) EncodeToJSONWithCode(code string) string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d,\"code\":\"%s\"}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix(), code)
	return encoded
}
