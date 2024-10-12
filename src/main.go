package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"faroe/ratelimit"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

var credential []byte

var db *sql.DB

var passwordHashingIPRateLimit = ratelimit.NewTokenBucketRateLimit(5, 10*time.Second)

var loginIPRateLimit = ratelimit.NewExpiringTokenBucketRateLimit(5, 10*time.Minute)

var createEmailVerificationUserRateLimit = ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute)

var verifyEmailVerificationCodeLimitCounter = ratelimit.NewLimitCounter(5)

var createPasswordResetIPRateLimit = ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute)

var verifyPasswordResetCodeLimitCounter = ratelimit.NewLimitCounter(5)

var totpUserRateLimit = ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute)

var recoveryCodeUserRateLimit = ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute)

//go:embed schema.sql
var schema string

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please define a port: faroe 3000")
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid port")
	}

	envKV := os.Environ()
	for _, kv := range envKV {
		parts := strings.Split(kv, "=")
		if len(parts) == 2 && parts[0] == "credential" {
			credential = []byte(parts[1])
		}
	}

	if len(credential) == 0 {
		log.Fatal("Please define a credential with --credential=CREDENTIAL")
	}

	go func() {
		for range time.Tick(10 * 24 * time.Hour) {
			passwordHashingIPRateLimit.Clear()
			loginIPRateLimit.Clear()
			createEmailVerificationUserRateLimit.Clear()
			verifyEmailVerificationCodeLimitCounter.Clear()
			createPasswordResetIPRateLimit.Clear()
			verifyPasswordResetCodeLimitCounter.Clear()
			totpUserRateLimit.Clear()
			log.Println("SYSTEM RESET_MEMORY_STORAGE")
		}
	}()

	err = os.MkdirAll("faroe_data", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("sqlite3", "./faroe_data/sqlite.db?_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		// for range time.Tick(1 * time.Hour) {
		// 	err = backupDatabase()
		// 	if err != nil {
		// 		log.Printf("SYSTEM DATABASE_BACKUP %e\n", err)
		// 	} else {
		// 		log.Println("SYSTEM DATABASE_BACKUP SUCCESS")
		// 	}
		// }
	}()

	router := &httprouter.Router{
		NotFound: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !verifyCredential(r) {
				writeNotAuthenticatedErrorResponse(w)
			} else {
				writeNotFoundErrorResponse(w)
			}
		}),
	}

	router.POST("/authenticate/password", handleAuthenticateWithPasswordRequest)

	router.POST("/users", handleCreateUserRequest)
	router.GET("/users", handleGetUsersRequest)
	router.GET("/users/:user_id", handleGetUserRequest)
	router.DELETE("/users/:user_id", handleDeleteUserRequest)
	router.POST("/users/:user_id/password", handleUpdatePasswordRequest)
	router.GET("/users/:user_id/totp", handleGetUserTOTPCredentialRequest)
	router.POST("/users/:user_id/totp", handleRegisterTOTPRequest)
	router.POST("/users/:user_id/verify-2fa/totp", handleVerifyTOTPRequest)
	router.POST("/users/:user_id/reset-2fa", handleResetUser2FARequest)
	router.POST("/users/:user_id/regenerate-recovery-code", handleRegenerateUserRecoveryCodeRequest)
	router.POST("/users/:user_id/email-verification", handleCreateEmailVerificationRequestRequest)
	router.GET("/users/:user_id/email-verification/:request_id", handleGetEmailVerificationRequestRequest)
	router.DELETE("/users/:user_id/email-verification/:request_id", handleDeleteEmailVerificationRequestRequest)
	router.POST("/users/:user_id/verify-email", handleVerifyUserEmailRequest)
	router.POST("/password-reset", handleCreatePasswordResetRequestRequest)
	router.GET("/password-reset/:request_id", handleGetPasswordResetRequestRequest)
	router.DELETE("/password-reset/:request_id", handleDeleteEmailVerificationRequestRequest)
	router.GET("/password-reset/:request_id/user", handleGetPasswordResetRequestUserRequest)
	router.POST("/password-reset/:request_id/verify-email", handleVerifyPasswordResetRequestEmailRequest)
	router.POST("/password-reset/:request_id/verify-2fa/totp", handleVerifyPasswordResetRequest2FAWithTOTPRequest)
	router.POST("/password-reset/:request_id/reset-2fa", handleResetPasswordResetRequest2FAWithRecoveryCodeRequest)
	router.POST("/reset-password", handleResetPasswordRequest)

	fmt.Printf("Starting server in port %d...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	log.Println(err)
}

func writeExpectedErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	escapedMessage := strings.ReplaceAll(string(message), "\"", "\\\"")
	w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", escapedMessage)))
}

func writeUnexpectedErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	w.Write([]byte("{\"error\":\"UNEXPECTED_ERROR\"}"))
}

func writeNotFoundErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(404)
	w.Write([]byte("{\"error\":\"NOT_FOUND\"}"))
}

func writeNotAcceptableErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(406)
	w.Write([]byte("{\"error\":\"NOT_ACCEPTABLE\"}"))
}

func writeUnsupportedMediaTypeErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(415)
	w.Write([]byte("{\"error\":\"UNSUPPORTED_MEDIA_TYPE\"}"))
}

func writeNotAuthenticatedErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)
	w.Write([]byte("{\"error\":\"NOT_AUTHENTICATED\"}"))
}

func logMessageWithClientIP(level string, action string, result string, clientIP string, attributes string) {
	if clientIP == "" {
		clientIP = "-"
	}
	log.Printf("%s %s %s client_ip=%s %s\n", level, action, result, clientIP, attributes)
}

func generateId() (string, error) {
	bytes := make([]byte, 15)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	id := base32.NewEncoding("abcdefghijkmnpqrstuvwxyz23456789").EncodeToString(bytes)
	return id, nil
}

const (
	expectedErrorInvalidData             = "INVALID_DATA"
	expectedErrorTooManyRequests         = "TOO_MANY_REQUESTS"
	expectedErrorInvalidEmail            = "INVALID_EMAIL"
	expectedErrorWeakPassword            = "WEAK_PASSWORD"
	expectedErrorPasswordTooLarge        = "PASSWORD_TOO_LARGE"
	expectedErrorEmailAlreadyUsed        = "EMAIL_ALREADY_USED"
	expectedErrorAccountNotExists        = "ACCOUNT_NOT_EXISTS"
	expectedErrorIncorrectPassword       = "INCORRECT_PASSWORD"
	expectedErrorIncorrectCode           = "INCORRECT_CODE"
	expectedErrorEmailNotVerified        = "EMAIL_NOT_VERIFIED"
	expectedErrorSecondFactorNotVerified = "SECOND_FACTOR_NOT_VERIFIED"
	expectedErrorSecondFactorNotAllowed  = "SECOND_FACTOR_NOT_ALLOWED"
	expectedErrorInvalidRequestId        = "INVALID_REQUEST_ID"
)

var ErrRecordNotFound = errors.New("record not found")
