package main

import (
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

func handleRegisterTOTPRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifySecret(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
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

	if data.Code == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	code := *data.Code
	if len(code) != 6 {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	validCode := otp.VerifyTOTP(key, 30*time.Second, 6, code)
	if !validCode {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	err = registerUserTOTPCredential(userId, key)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	recoveryCode, err := getUserRecoveryCode(userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(encodeRecoveryCodeToJSON(recoveryCode)))
}

func handleVerifyTOTPRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	credential, err := getUserTOTPCredential(userId)
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
	if data.Code == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if !totpUserRateLimit.Consume(userId, 1) {
		logMessageWithClientIP("INFO", "VERIFY_2FA_TOTP", "TOTP_USER_LIMIT_REJECTED", clientIP, fmt.Sprintf("user_id=%s", userId))
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	valid := otp.VerifyTOTP(credential.Key, 30*time.Second, 6, *data.Code)
	if !valid {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	totpUserRateLimit.Reset(userId)
	w.WriteHeader(204)
}

func handleDeleteUserTOTPCredentialRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifySecret(r) {
		writeNotAuthenticatedErrorResponse(w)
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

	err = deleteUserTOTPCredential(userId)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetUserTOTPCredentialRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifySecret(r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	userId := params.ByName("user_id")
	credential, err := getUserTOTPCredential(userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(credential.EncodeToJSON()))
}

func getUserTOTPCredential(userId string) (TOTPCredential, error) {
	var credential TOTPCredential
	var createdAtUnix int64
	row := db.QueryRow("SELECT id, user_id, created_at, key FROM totp_credential WHERE user_id = ?", userId)
	err := row.Scan(&credential.Id, &credential.UserId, &createdAtUnix, &credential.Key)
	if errors.Is(err, sql.ErrNoRows) {
		return TOTPCredential{}, ErrRecordNotFound
	}
	credential.CreatedAt = time.Unix(createdAtUnix, 0)
	return credential, nil
}

func registerUserTOTPCredential(userId string, key []byte) error {
	now := time.Now()
	id, err := generateId()
	if err != nil {
		return nil
	}
	result, err := db.Exec(`INSERT INTO totp_credential (id, user_id, created_at, key) VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET created_at = ?, key = ? WHERE user_id = ?`, id, userId, now.Unix(), key, now.Unix(), key, userId)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count < 1 {
		return ErrRecordNotFound
	}
	return nil
}

func deleteUserTOTPCredential(userId string) error {
	_, err := db.Exec("DELETE FROM totp_credential WHERE user_id = ?", userId)
	return err
}

type TOTPCredential struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	Key       []byte
}

func (c *TOTPCredential) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"key\":\"%s\"}", c.Id, c.UserId, c.CreatedAt.Unix(), base64.StdEncoding.EncodeToString(c.Key))
	return encoded
}
