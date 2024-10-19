package main

import (
	"context"
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

func handleCreateUserEmailVerificationRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	if !env.createEmailVerificationUserRateLimit.Consume(userId, 1) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	now := time.Now()
	requestId, err := generateId()
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	code, err := generateSecureCode()
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	verificationRequest := EmailVerificationRequest{
		Id:        requestId,
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		Code:      code,
	}

	err = createEmailVerificationRequest(env.db, r.Context(), verificationRequest)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

func handleVerifyUserEmailRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		env.createEmailVerificationUserRateLimit.AddToken(userId, 1)
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
		err = deleteEmailVerificationRequest(env.db, r.Context(), verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		env.createEmailVerificationUserRateLimit.AddToken(userId, 1)
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

	if !env.verifyUserEmailRateLimit.Consume(userId, 1) {
		err = deleteEmailVerificationRequest(env.db, r.Context(), verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := validateUserEmailVerificationRequest(env.db, r.Context(), userId, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validCode {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	env.verifyUserEmailRateLimit.Reset(verificationRequest.UserId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}

func handleDeleteUserEmailVerificationRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	err = deleteEmailVerificationRequest(env.db, r.Context(), verificationRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func handleGetUserEmailVerificationRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
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
		err = deleteEmailVerificationRequest(env.db, r.Context(), verificationRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

func createEmailVerificationRequest(db *sql.DB, ctx context.Context, request EmailVerificationRequest) error {
	_, err := db.ExecContext(ctx, `INSERT INTO email_verification_request (id, user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET id = ?, created_at = ?, code = ? WHERE user_id = ?`, request.Id, request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code, request.Id, request.CreatedAt.Unix(), request.Code, request.UserId)
	return err
}

func getUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string) (EmailVerificationRequest, error) {
	var verificationRequest EmailVerificationRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT id, user_id, created_at, expires_at, code FROM email_verification_request WHERE user_id = ?", userId)
	err := row.Scan(&verificationRequest.Id, &verificationRequest.UserId, &createdAtUnix, &expiresAtUnix, &verificationRequest.Code)
	if errors.Is(err, sql.ErrNoRows) {
		return EmailVerificationRequest{}, ErrRecordNotFound
	}
	verificationRequest.CreatedAt = time.Unix(createdAtUnix, 0)
	verificationRequest.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return verificationRequest, nil
}

func deleteUserEmailVerificationRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_verification_request WHERE user_id = ?", userId)
	return err
}

func deleteEmailVerificationRequest(db *sql.DB, ctx context.Context, requestId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_verification_request WHERE id = ?", requestId)
	return err
}

func validateUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string, code string) (bool, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM email_verification_request WHERE user_id = ? AND code = ? AND expires_at > ?", userId, code, time.Now().Unix())
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
