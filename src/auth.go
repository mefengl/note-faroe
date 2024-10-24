package main

import (
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func handleAuthenticateWithPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	email, password := strings.ToLower(*data.Email), *data.Password
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
	if clientIP != "" && !env.passwordHashingIPRateLimit.Consume(clientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if clientIP != "" && !env.loginIPRateLimit.Consume(clientIP) {
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
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(user.EncodeToJSON()))
}
