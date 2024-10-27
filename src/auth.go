package main

import (
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func handleVerifyUserPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	user, err := getUser(env.db, r.Context(), userId)
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
		Password *string `json:"password"`
		ClientIP string  `json:"client_ip"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	if data.ClientIP != "" && !env.loginIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	validPassword, err := argon2id.Verify(user.PasswordHash, *data.Password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !validPassword {
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}
	if data.ClientIP != "" {
		env.loginIPRateLimit.AddTokenIfEmpty(data.ClientIP)
	}
	w.WriteHeader(204)
}
