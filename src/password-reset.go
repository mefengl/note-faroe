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

func handleCreateUserPasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	if len(body) > 0 {
		var data struct {
			ClientIP string `json:"client_ip"`
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
			return
		}

		if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
			writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
			return
		}
		if data.ClientIP != "" && !env.createPasswordResetIPRateLimit.Consume(data.ClientIP) {
			writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
			return
		}
	}

	err = deleteExpiredUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	code, err := generateSecureCode()
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

	resetRequest, err := createPasswordResetRequest(env.db, r.Context(), userId, codeHash)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
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
	w.Write([]byte(resetRequest.EncodeToJSON()))
}

func handleVerifyPasswordResetRequestEmailRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
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
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	var data struct {
		Code     *string `json:"code"`
		ClientIP string  `json:"client_ip"`
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
	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if !env.verifyPasswordResetCodeLimitCounter.Consume(resetRequest.Id) {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validCode, err := argon2id.Verify(resetRequest.CodeHash, *data.Code)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}
	var data struct {
		RequestId *string `json:"request_id"`
		Password  *string `json:"password"`
		ClientIP  string  `json:"client_ip"`
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

	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), *data.RequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}
	if err != nil {
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}

	password := *data.Password
	if len(password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	strongPassword, err := verifyPasswordStrength(password)
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
	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	validResetRequest, err := resetUserPasswordWithPasswordResetRequest(env.db, r.Context(), resetRequest.Id, passwordHash)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetUserPasswordResetRequestsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	err = deleteExpiredUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	resetRequest, err := getUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if len(resetRequest) == 0 {
		w.Write([]byte("[]"))
		return
	}
	w.Write([]byte("["))
	for i, user := range resetRequest {
		w.Write([]byte(user.EncodeToJSON()))
		if i != len(resetRequest)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))
}

func handleDeleteUserPasswordResetRequestsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	err = deleteUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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

func getUserPasswordResetRequests(db *sql.DB, ctx context.Context, requestId string) ([]PasswordResetRequest, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, created_at, code_hash, expires_at FROM password_reset_request WHERE id = ?", requestId)
	if err != nil {
		return nil, err
	}
	var requests []PasswordResetRequest
	defer rows.Close()
	for rows.Next() {
		var request PasswordResetRequest
		var createdAtUnix, expiresAtUnix int64
		err := rows.Scan(&request.Id, &request.UserId, &createdAtUnix, &request.CodeHash, &expiresAtUnix)
		if err != nil {
			return nil, err
		}
		request.CreatedAt = time.Unix(createdAtUnix, 0)
		request.ExpiresAt = time.Unix(expiresAtUnix, 0)
		requests = append(requests, request)
	}
	return requests, nil
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

func deleteExpiredUserPasswordResetRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM password_reset_request WHERE user_id = ? AND expires_at <= ?", userId, time.Now().Unix())
	return err
}

func deleteUserPasswordResetRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM password_reset_request WHERE user_id = ?", userId)
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
