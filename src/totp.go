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

func handleRegisterUserTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
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

	credential, err := createTOTPCredential(env.db, r.Context(), userId, key)
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
	w.Write([]byte(credential.EncodeToJSON()))
}

func handleVerifyTOTPCredentialTOTPRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	credentialId := params.ByName("credential_id")
	credential, err := getTOTPCredential(env.db, r.Context(), credentialId)
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
	if !env.totpUserRateLimit.Consume(credential.UserId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	valid := otp.VerifyTOTPWithGracePeriod(time.Now(), credential.Key, 30*time.Second, 6, *data.Code, 10*time.Second)
	if !valid {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	env.totpUserRateLimit.Reset(credential.UserId)

	w.WriteHeader(204)
}

func handleDeleteTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	credentialId := params.ByName("credential_id")
	_, err := getTOTPCredential(env.db, r.Context(), credentialId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	err = deleteTOTPCredential(env.db, r.Context(), credentialId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	credentialId := params.ByName("credential_id")
	credential, err := getTOTPCredential(env.db, r.Context(), credentialId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(credential.EncodeToJSON()))
}

func handleGetUserTOTPCredentialsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	credentials, err := getUserTOTPCredentials(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if len(credentials) == 0 {
		w.Write([]byte("[]"))
		return
	}
	w.Write([]byte("["))
	for i, credential := range credentials {
		w.Write([]byte(credential.EncodeToJSON()))
		if i != len(credentials)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))
}

func handleDeleteUserTOTPCredentialRequests(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	err = deleteUserTOTPCredentials(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

func getTOTPCredential(db *sql.DB, ctx context.Context, credentialId string) (TOTPCredential, error) {
	var credential TOTPCredential
	var createdAtUnix int64
	row := db.QueryRowContext(ctx, "SELECT id, user_id, created_at, key FROM totp_credential WHERE id = ?", credentialId)
	err := row.Scan(&credential.Id, &credential.UserId, &createdAtUnix, &credential.Key)
	if errors.Is(err, sql.ErrNoRows) {
		return TOTPCredential{}, ErrRecordNotFound
	}
	credential.CreatedAt = time.Unix(createdAtUnix, 0)
	return credential, nil
}

func createTOTPCredential(db *sql.DB, ctx context.Context, userId string, key []byte) (TOTPCredential, error) {
	id, err := generateId()
	if err != nil {
		return TOTPCredential{}, nil
	}
	credential := TOTPCredential{
		Id:        id,
		UserId:    userId,
		CreatedAt: time.Unix(time.Now().Unix(), 0),
		Key:       key,
	}
	err = insertTOTPCredential(db, ctx, &credential)
	if err != nil {
		return TOTPCredential{}, err
	}
	return credential, nil
}

func insertTOTPCredential(db *sql.DB, ctx context.Context, credential *TOTPCredential) error {
	_, err := db.ExecContext(ctx, "INSERT INTO totp_credential (id, user_id, created_at, key) VALUES (?, ?, ?, ?)", credential.Id, credential.UserId, credential.CreatedAt.Unix(), credential.Key)
	return err
}

func deleteTOTPCredential(db *sql.DB, ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM totp_credential WHERE id = ?", id)
	return err
}

func getUserTOTPCredentials(db *sql.DB, ctx context.Context, userId string) ([]TOTPCredential, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, created_at, key FROM totp_credential WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}
	var credentials []TOTPCredential
	defer rows.Close()
	for rows.Next() {
		var credential TOTPCredential
		var createdAtUnix int64
		err := rows.Scan(&credential.Id, &credential.UserId, &createdAtUnix, &credential.Key)
		if err != nil {
			return nil, err
		}
		credential.CreatedAt = time.Unix(createdAtUnix, 0)
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func deleteUserTOTPCredentials(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM totp_credential WHERE user_id = ?", userId)
	return err
}

type TOTPCredential struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	Key       []byte
}

func (c *TOTPCredential) EncodeToJSON() string {
	encoded := fmt.Sprintf(`{"id":"%s","user_id":"%s","created_at":%d,"key":"%s"}`, c.Id, c.UserId, c.CreatedAt.Unix(), base64.StdEncoding.EncodeToString(c.Key))
	return encoded
}
