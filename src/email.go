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
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleCreateUserEmailUpdateRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	emailAvailable, err := checkEmailAvailability(env.db, r.Context(), email)
	if !emailAvailable {
		writeExpectedErrorResponse(w, ExpectedErrorEmailAlreadyUsed)
		return
	}

	if !env.createEmailVerificationUserRateLimit.Consume(userId, 1) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	updateRequest, err := createEmailUpdateRequest(env.db, r.Context(), userId, email)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(updateRequest.EncodeToJSON()))
}

func handleUpdateEmailRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	updateRequest, err := getEmailUpdateRequest(env.db, r.Context(), *data.RequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	// If now is or after expiration
	if time.Now().Compare(updateRequest.ExpiresAt) >= 0 {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}
	if !env.verifyEmailUpdateVerificationCodeLimitCounter.Consume(updateRequest.Id) {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnExpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := validateEmailUpdateRequest(env.db, r.Context(), updateRequest.Id, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validCode {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	env.verifyEmailUpdateVerificationCodeLimitCounter.Delete(updateRequest.Id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(encodeEmailToJSON(updateRequest.Email)))
}

func handleDeleteEmailUpdateRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	updateRequestId := params.ByName("request_id")
	updateRequest, err := getEmailUpdateRequest(env.db, r.Context(), updateRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetEmailUpdateRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	updateRequestId := params.ByName("request_id")
	updateRequest, err := getEmailUpdateRequest(env.db, r.Context(), updateRequestId)
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
	if time.Now().Compare(updateRequest.ExpiresAt) >= 0 {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
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
	w.Write([]byte(updateRequest.EncodeToJSON()))
}

func createEmailUpdateRequest(db *sql.DB, ctx context.Context, userId string, email string) (EmailUpdateRequest, error) {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return EmailUpdateRequest{}, nil
	}
	expiresAt := now.Add(10 * time.Minute)
	code, err := generateSecureCode()
	if err != nil {
		return EmailUpdateRequest{}, nil
	}
	_, err = db.ExecContext(ctx, "INSERT INTO email_update_request (id, user_id, created_at, expires_at, email, code) VALUES (?, ?, ?, ?, ?, ?)", id, userId, now.Unix(), expiresAt.Unix(), email, code)
	if err != nil {
		return EmailUpdateRequest{}, err
	}
	request := EmailUpdateRequest{
		Id:        id,
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Email:     email,
		Code:      code,
	}
	return request, nil
}

func getEmailUpdateRequest(db *sql.DB, ctx context.Context, requestId string) (EmailUpdateRequest, error) {
	var updateRequest EmailUpdateRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT id, user_id, created_at, email, code, expires_at FROM email_update_request WHERE id = ?", requestId)
	err := row.Scan(&updateRequest.Id, &updateRequest.UserId, &createdAtUnix, &updateRequest.Email, &updateRequest.Code, &expiresAtUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return EmailUpdateRequest{}, ErrRecordNotFound
	}
	updateRequest.CreatedAt = time.Unix(createdAtUnix, 0)
	updateRequest.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return updateRequest, nil
}

func deleteUserEmailUpdateRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_update_request WHERE user_id = ?", userId)
	return err
}

func deleteEmailUpdateRequest(db *sql.DB, ctx context.Context, requestId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_update_request WHERE id = ?", requestId)
	return err
}

func validateEmailUpdateRequest(db *sql.DB, ctx context.Context, requestId string, code string) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	row := tx.QueryRow("DELETE FROM email_update_request WHERE id = ? AND code = ? AND expires_at > ? RETURNING user_id, email", requestId, code, time.Now().Unix())
	var userId, email string
	err = row.Scan(&userId, &email)
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
	_, err = tx.Exec("UPDATE user SET email = ? WHERE id = ?", email, userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	_, err = tx.Exec("DELETE FROM password_reset_request WHERE user_id = ?", userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	_, err = tx.Exec("DELETE FROM email_update_request WHERE email = ?", email)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return false, err
	}
	tx.Commit()
	return true, nil
}

type EmailUpdateRequest struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	Email     string
	Code      string
	ExpiresAt time.Time
}

func (r *EmailUpdateRequest) EncodeToJSON() string {
	escapedEmail := strings.ReplaceAll(r.Email, "\"", "\\\"")
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"email\":\"%s\",\"expires_at\":%d,\"code\":\"%s\"}", r.Id, r.UserId, r.CreatedAt.Unix(), escapedEmail, r.ExpiresAt.Unix(), r.Code)
	return encoded
}

func encodeEmailToJSON(email string) string {
	escapedEmail := strings.ReplaceAll(email, "\"", "\\\"")
	encoded := fmt.Sprintf("{\"email\":\"%s\"}", escapedEmail)
	return encoded
}
