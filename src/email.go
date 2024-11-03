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
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	if !env.verifyUserEmailRateLimit.Check(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if !env.createEmailRequestUserRateLimit.Consume(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	verificationRequest, err := createUserEmailVerificationRequest(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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

	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		env.createEmailRequestUserRateLimit.AddTokenIfEmpty(userId)
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// If now is or after expiration
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		env.createEmailRequestUserRateLimit.AddTokenIfEmpty(userId)
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
	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if !env.verifyUserEmailRateLimit.Consume(userId) {
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := validateUserEmailVerificationRequest(env.db, r.Context(), userId, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}

	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}

	// If now is or after expiration
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

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
	email := strings.ToLower(*data.Email)
	if !verifyEmailInput(email) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if !env.createEmailRequestUserRateLimit.Consume(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	err = deleteExpiredUserEmailUpdateRequest(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	updateRequest, err := createEmailUpdateRequest(env.db, r.Context(), userId, email)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
	if data.Code == nil || *data.Code == "" {
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
		writeUnexpectedErrorResponse(w)
		return
	}

	// If now is or after expiration
	if time.Now().Compare(updateRequest.ExpiresAt) >= 0 {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}
	if !env.verifyEmailUpdateVerificationCodeLimitCounter.Consume(updateRequest.Id) {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := redeemUpdateRequest(env.db, r.Context(), updateRequest.Id, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(updateRequest.ExpiresAt) >= 0 {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
	}

	err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(updateRequest.ExpiresAt) >= 0 {
		err = deleteEmailUpdateRequest(env.db, r.Context(), updateRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(updateRequest.EncodeToJSON()))
}

func createUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string) (UserEmailVerificationRequest, error) {
	now := time.Unix(time.Now().Unix(), 0)
	code, err := generateSecureCode()
	if err != nil {
		return UserEmailVerificationRequest{}, err
	}
	request := UserEmailVerificationRequest{
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		Code:      code,
	}
	_, err = db.ExecContext(ctx, `INSERT INTO user_email_verification_request (user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?)
	ON CONFLICT (user_id) DO UPDATE SET created_at = ?, code = ? WHERE user_id = ?`, request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code, request.CreatedAt.Unix(), request.Code, request.UserId)
	if err != nil {
		return UserEmailVerificationRequest{}, err
	}
	return request, nil
}

func handleGetUserEmailUpdateRequestsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	err = deleteExpiredUserEmailUpdateRequest(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	updateRequests, err := getUserEmailUpdateRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if len(updateRequests) == 0 {
		w.Write([]byte("[]"))
		return
	}
	w.Write([]byte("["))
	for i, updateRequest := range updateRequests {
		w.Write([]byte(updateRequest.EncodeToJSON()))
		if i != len(updateRequests)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))
}

func handleDeleteUserEmailUpdateRequestsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	err = deleteUserEmailUpdateRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func getUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string) (UserEmailVerificationRequest, error) {
	var verificationRequest UserEmailVerificationRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT user_id, created_at, expires_at, code FROM user_email_verification_request WHERE user_id = ?", userId)
	err := row.Scan(&verificationRequest.UserId, &createdAtUnix, &expiresAtUnix, &verificationRequest.Code)
	if errors.Is(err, sql.ErrNoRows) {
		return UserEmailVerificationRequest{}, ErrRecordNotFound
	}
	verificationRequest.CreatedAt = time.Unix(createdAtUnix, 0)
	verificationRequest.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return verificationRequest, nil
}

func deleteUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM user_email_verification_request WHERE user_id = ?", userId)
	return err
}

func validateUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string, code string) (bool, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM user_email_verification_request WHERE user_id = ? AND code = ? AND expires_at > ?", userId, code, time.Now().Unix())
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

type UserEmailVerificationRequest struct {
	UserId    string
	CreatedAt time.Time
	Code      string
	ExpiresAt time.Time
}

func (r *UserEmailVerificationRequest) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d,\"code\":\"%s\"}", r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix(), r.Code)
	return encoded
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
	request := EmailUpdateRequest{
		Id:        id,
		UserId:    userId,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Email:     email,
		Code:      code,
	}
	err = insertEmailUpdateRequest(db, ctx, &request)
	if err != nil {
		return EmailUpdateRequest{}, err
	}
	return request, nil
}

func insertEmailUpdateRequest(db *sql.DB, ctx context.Context, request *EmailUpdateRequest) error {
	_, err := db.ExecContext(ctx, "INSERT INTO email_update_request (id, user_id, created_at, expires_at, email, code) VALUES (?, ?, ?, ?, ?, ?)", request.Id, request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Email, request.Code)
	return err
}

func getEmailUpdateRequest(db *sql.DB, ctx context.Context, requestId string) (EmailUpdateRequest, error) {
	var request EmailUpdateRequest
	var createdAtUnix, expiresAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT id, user_id, created_at, email, code, expires_at FROM email_update_request WHERE id = ?", requestId)
	err := row.Scan(&request.Id, &request.UserId, &createdAtUnix, &request.Email, &request.Code, &expiresAtUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return EmailUpdateRequest{}, ErrRecordNotFound
	}
	request.CreatedAt = time.Unix(createdAtUnix, 0)
	request.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return request, nil
}

func getUserEmailUpdateRequests(db *sql.DB, ctx context.Context, userId string) ([]EmailUpdateRequest, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, created_at, email, code, expires_at FROM email_update_request WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}
	var requests []EmailUpdateRequest
	defer rows.Close()
	for rows.Next() {
		var request EmailUpdateRequest
		var createdAtUnix, expiresAtUnix int64
		err := rows.Scan(&request.Id, &request.UserId, &createdAtUnix, &request.Email, &request.Code, &expiresAtUnix)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		request.CreatedAt = time.Unix(createdAtUnix, 0)
		request.ExpiresAt = time.Unix(expiresAtUnix, 0)
		requests = append(requests, request)
	}
	return requests, nil
}

func deleteUserEmailUpdateRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_update_request WHERE user_id = ?", userId)
	return err
}

func deleteEmailUpdateRequest(db *sql.DB, ctx context.Context, requestId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_update_request WHERE id = ?", requestId)
	return err
}

func deleteExpiredUserEmailUpdateRequest(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM email_update_request WHERE user_id = ? AND expires_at <= ?", userId, time.Now().Unix())
	return err
}

func redeemUpdateRequest(db *sql.DB, ctx context.Context, requestId string, code string) (bool, error) {
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
