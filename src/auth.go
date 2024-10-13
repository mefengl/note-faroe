package main

import (
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func handleAuthenticateWithPasswordRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	var data struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if data.Email == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	email, password := *data.Email, *data.Password
	user, err := getUserFromEmail(email)
	if errors.Is(err, ErrRecordNotFound) {
		logMessageWithClientIP("INFO", "LOGIN_ATTEMPT", "INVALID_EMAIL", clientIP, fmt.Sprintf("email_input=\"%s\"", strings.ReplaceAll(email, "\"", "\\\"")))
		writeExpectedErrorResponse(w, ExpectedErrorAccountNotExists)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if clientIP != "" && !passwordHashingIPRateLimit.Consume(clientIP, 1) {
		logMessageWithClientIP("INFO", "AUTHENTICATE_WITH_PASSWORD", "PASSWORD_HASHING_LIMIT_REJECTED", clientIP, "")
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if clientIP != "" && !loginIPRateLimit.Consume(clientIP, 1) {
		logMessageWithClientIP("INFO", "AUTHENTICATE_WITH_PASSWORD", "LOGIN_LIMIT_REJECTED", clientIP, "")
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validPassword, err := argon2id.Verify(user.PasswordHash, password)
	if err != nil {
		log.Println(err)
		writeUnExpectedErrorResponse(w)
		return
	}
	if !validPassword {
		logMessageWithClientIP("INFO", "LOGIN_ATTEMPT", "INVALID_PASSWORD", clientIP, fmt.Sprintf("email=\"%s\" user_id=%s", strings.ReplaceAll(email, "\"", "\\\""), user.Id))
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}
	if clientIP != "" {
		loginIPRateLimit.AddToken(clientIP, 1)
	}
	logMessageWithClientIP("INFO", "LOGIN_ATTEMPT", "SUCCESS", clientIP, fmt.Sprintf("email=\"%s\" user_id=%s", strings.ReplaceAll(email, "\"", "\\\""), user.Id))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}
