package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"faroe/otp"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

func handleRegisterTOTPRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
		Key  *string `json:"key"`
		Code *string `json:"code"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.Key == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	key, err := base64.StdEncoding.DecodeString(*data.Key)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if len(key) != 20 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	validCode := otp.VerifyTOTPWithGracePeriod(time.Now(), key, 30*time.Second, 6, *data.Code, 10*time.Second)
	if !validCode {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	credential, err := registerUserTOTPCredential(env.db, r.Context(), userId, key)
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
	w.Write([]byte(credential.EncodeToJSON()))
}

func handleVerifyTOTPRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	credential, err := getUserTOTPCredential(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
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
	if !env.totpUserRateLimit.Consume(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	valid := otp.VerifyTOTPWithGracePeriod(time.Now(), credential.Key, 30*time.Second, 6, *data.Code, 10*time.Second)
	if !valid {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	env.totpUserRateLimit.Reset(userId)

	w.WriteHeader(204)
}

func handleDeleteUserTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
	_, err := getUserTOTPCredential(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	err = deleteUserTOTPCredential(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetUserTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	userId := params.ByName("user_id")
	credential, err := getUserTOTPCredential(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(credential.EncodeToJSON()))
}

func getUserTOTPCredential(db *sql.DB, ctx context.Context, userId string) (UserTOTPCredential, error) {
	var credential UserTOTPCredential
	var createdAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT user_id, created_at, key FROM user_totp_credential WHERE user_id = ?", userId)
	err := row.Scan(&credential.UserId, &createdAtUnix, &credential.Key)
	if errors.Is(err, sql.ErrNoRows) {
		return UserTOTPCredential{}, ErrRecordNotFound
	}
	credential.CreatedAt = time.Unix(createdAtUnix, 0)
	return credential, nil
}

func registerUserTOTPCredential(db *sql.DB, ctx context.Context, userId string, key []byte) (UserTOTPCredential, error) {
	credential := UserTOTPCredential{
		UserId:    userId,
		CreatedAt: time.Unix(time.Now().Unix(), 0),
		Key:       key,
	}
	_, err := db.ExecContext(ctx, `INSERT INTO user_totp_credential (user_id, created_at, key) VALUES (?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET created_at = ?, key = ? WHERE user_id = ?`, credential.UserId, credential.CreatedAt.Unix(), credential.Key, credential.CreatedAt.Unix(), credential.Key, credential.UserId)
	if err != nil {
		return UserTOTPCredential{}, err
	}
	return credential, nil
}

func deleteUserTOTPCredential(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM user_totp_credential WHERE user_id = ?", userId)
	return err
}

type UserTOTPCredential struct {
	UserId    string
	CreatedAt time.Time
	Key       []byte
}

func (c *UserTOTPCredential) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"user_id\":\"%s\",\"created_at\":%d,\"key\":\"%s\"}", c.UserId, c.CreatedAt.Unix(), base64.StdEncoding.EncodeToString(c.Key))
	return encoded
}
