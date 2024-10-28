package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"faroe/ratelimit"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	_ "embed"

	"github.com/julienschmidt/httprouter"
	_ "modernc.org/sqlite"
)

const version = "0.1.0"

//go:embed schema.sql
var schema string

func main() {
	for _, arg := range os.Args {
		if arg == "-v" || arg == "--version" {
			fmt.Printf("Faroe version %s\n", version)
			return
		}
	}

	if len(os.Args) < 2 {
		fmt.Print(`Usage:

faroe serve - Start the Faroe server
faroe generate-secret - Generate a secure secret

`)
		return
	}

	if os.Args[1] == "serve" {
		serveCommand()
		return
	}
	if os.Args[1] == "generate-secret" {
		generateSecretCommand()
		return
	}

	flag.Parse()
	fmt.Println("Unknown command")
}

func generateSecretCommand() {
	// Remove "server" command since Go's flag package stops parsing at first non-flag argument.
	os.Args = os.Args[1:]
	flag.Parse()

	bytes := make([]byte, 25)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal("Failed to generate secret\n")
	}
	secret := base32.NewEncoding("abcdefghijkmnpqrstuvwxyz23456789").EncodeToString(bytes)
	fmt.Println(secret)
}

func serveCommand() {
	// Remove "server" command since Go's flag package stops parsing at first non-flag argument.
	os.Args = os.Args[1:]

	var port int
	var dataDir string
	var secretString string
	flag.IntVar(&port, "port", 4000, "Port number")
	flag.StringVar(&dataDir, "dir", "faroe_data", "Data directory name")
	flag.StringVar(&secretString, "secret", "", "Server secret")
	flag.Parse()

	secret := []byte(secretString)

	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", path.Join(dataDir, "sqlite.db"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for range time.Tick(10 * 24 * time.Hour) {
			err := cleanUpDatabase(db)
			if err != nil {
				log.Println(err)
			}
		}
	}()

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

	passwordHashingIPRateLimit := ratelimit.NewTokenBucketRateLimit(5, 10*time.Second)
	loginIPRateLimit := ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute)
	createEmailRequestUserRateLimit := ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute)
	verifyUserEmailRateLimit := ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute)
	verifyEmailUpdateVerificationCodeLimitCounter := ratelimit.NewLimitCounter(5)
	createPasswordResetIPRateLimit := ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute)
	verifyPasswordResetCodeLimitCounter := ratelimit.NewLimitCounter(5)
	totpUserRateLimit := ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute)
	recoveryCodeUserRateLimit := ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute)

	go func() {
		for range time.Tick(10 * 24 * time.Hour) {
			passwordHashingIPRateLimit.Clear()
			loginIPRateLimit.Clear()
			createEmailRequestUserRateLimit.Clear()
			verifyUserEmailRateLimit.Clear()
			verifyEmailUpdateVerificationCodeLimitCounter.Clear()
			createPasswordResetIPRateLimit.Clear()
			verifyPasswordResetCodeLimitCounter.Clear()
			totpUserRateLimit.Clear()
			recoveryCodeUserRateLimit.Clear()
			log.Println("SYSTEM RESET_MEMORY_STORAGE")
		}
	}()

	env := &Environment{
		db:                              db,
		secret:                          secret,
		passwordHashingIPRateLimit:      passwordHashingIPRateLimit,
		loginIPRateLimit:                loginIPRateLimit,
		createEmailRequestUserRateLimit: createEmailRequestUserRateLimit,
		verifyUserEmailRateLimit:        verifyUserEmailRateLimit,
		verifyEmailUpdateVerificationCodeLimitCounter: verifyEmailUpdateVerificationCodeLimitCounter,
		createPasswordResetIPRateLimit:                createPasswordResetIPRateLimit,
		verifyPasswordResetCodeLimitCounter:           verifyPasswordResetCodeLimitCounter,
		totpUserRateLimit:                             totpUserRateLimit,
		recoveryCodeUserRateLimit:                     recoveryCodeUserRateLimit,
	}

	app := CreateApp(env)
	fmt.Printf("Starting server in port %d...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app)

	log.Println(err)
}

func writeExpectedErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	escapedMessage := strings.ReplaceAll(string(message), "\"", "\\\"")
	w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", escapedMessage)))
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

func generateId() (string, error) {
	bytes := make([]byte, 15)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	id := base32.NewEncoding("abcdefghijkmnpqrstuvwxyz23456789").EncodeToString(bytes)
	return id, nil
}

type SortOrder int

const (
	SortOrderAscending SortOrder = iota
	SortOrderDescending
)

const (
	ExpectedErrorInvalidData             = "INVALID_DATA"
	ExpectedErrorTooManyRequests         = "TOO_MANY_REQUESTS"
	ExpectedErrorWeakPassword            = "WEAK_PASSWORD"
	ExpectedErrorUserNotExists           = "USER_NOT_EXISTS"
	ExpectedErrorIncorrectPassword       = "INCORRECT_PASSWORD"
	ExpectedErrorIncorrectCode           = "INCORRECT_CODE"
	ExpectedErrorSecondFactorNotVerified = "SECOND_FACTOR_NOT_VERIFIED"
	ExpectedErrorInvalidRequest          = "INVALID_REQUEST"
	ExpectedErrorNotAllowed              = "NOT_ALLOWED"
)

var ErrRecordNotFound = errors.New("record not found")

func NewRouter(env *Environment, defaultHandle RouteHandle) Router {
	router := Router{
		r: &httprouter.Router{
			NotFound: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defaultHandle(env, w, r, httprouter.Params{})
			}),
		},
		env: env,
	}
	return router
}

type Router struct {
	r   *httprouter.Router
	env *Environment
}

func (router *Router) Handle(method string, path string, handle RouteHandle) {
	router.r.Handle(method, path, func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		handle(router.env, w, r, params)
	})
}

func (router *Router) Handler() http.Handler {
	return router.r
}

type RouteHandle = func(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params)

type Environment struct {
	db                                            *sql.DB
	secret                                        []byte
	passwordHashingIPRateLimit                    ratelimit.TokenBucketRateLimit
	loginIPRateLimit                              ratelimit.ExpiringTokenBucketRateLimit
	createEmailRequestUserRateLimit               ratelimit.TokenBucketRateLimit
	verifyUserEmailRateLimit                      ratelimit.ExpiringTokenBucketRateLimit
	verifyEmailUpdateVerificationCodeLimitCounter ratelimit.LimitCounter
	createPasswordResetIPRateLimit                ratelimit.TokenBucketRateLimit
	verifyPasswordResetCodeLimitCounter           ratelimit.LimitCounter
	totpUserRateLimit                             ratelimit.ExpiringTokenBucketRateLimit
	recoveryCodeUserRateLimit                     ratelimit.ExpiringTokenBucketRateLimit
}

func CreateApp(env *Environment) http.Handler {
	router := NewRouter(env, func(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if !verifyRequestSecret(env.secret, r) {
			writeNotAuthenticatedErrorResponse(w)
		} else {
			writeNotFoundErrorResponse(w)
		}
	})

	router.Handle("GET", "/", func(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if !verifyRequestSecret(env.secret, r) {
			writeNotAuthenticatedErrorResponse(w)
			return
		}
		w.Write([]byte(fmt.Sprintf("Faroe version %s\n\nRead the documentation: https://faroe.dev\n", version)))
	})

	router.Handle("POST", "/users", handleCreateUserRequest)
	router.Handle("GET", "/users", handleGetUsersRequest)
	router.Handle("DELETE", "/users", handleDeleteUsersRequest)
	router.Handle("GET", "/users/:user_id", handleGetUserRequest)
	router.Handle("DELETE", "/users/:user_id", handleDeleteUserRequest)
	router.Handle("POST", "/users/:user_id/verify-password", handleVerifyUserPasswordRequest)
	router.Handle("POST", "/users/:user_id/update-password", handleUpdateUserPasswordRequest)
	router.Handle("POST", "/users/:user_id/register-totp", handleRegisterTOTPRequest)
	router.Handle("GET", "/users/:user_id/totp-credential", handleGetUserTOTPCredentialRequest)
	router.Handle("DELETE", "/users/:user_id/totp-credential", handleDeleteUserTOTPCredentialRequest)
	router.Handle("POST", "/users/:user_id/verify-2fa/totp", handleVerifyTOTPRequest)
	router.Handle("POST", "/users/:user_id/reset-2fa", handleResetUser2FARequest)
	router.Handle("POST", "/users/:user_id/regenerate-recovery-code", handleRegenerateUserRecoveryCodeRequest)
	router.Handle("POST", "/users/:user_id/email-verification-request", handleCreateUserEmailVerificationRequestRequest)
	router.Handle("GET", "/users/:user_id/email-verification-request", handleGetUserEmailVerificationRequestRequest)
	router.Handle("DELETE", "/users/:user_id/email-verification-request", handleDeleteUserEmailVerificationRequestRequest)
	router.Handle("POST", "/users/:user_id/verify-email", handleVerifyUserEmailRequest)
	router.Handle("POST", "/users/:user_id/email-update-requests", handleCreateUserEmailUpdateRequestRequest)
	router.Handle("GET", "/users/:user_id/email-update-requests", handleGetUserEmailUpdateRequestsRequest)
	router.Handle("DELETE", "/users/:user_id/email-update-requests", handleDeleteUserEmailUpdateRequestsRequest)
	router.Handle("GET", "/email-update-requests/:request_id", handleGetEmailUpdateRequestRequest)
	router.Handle("DELETE", "/email-update-requests/:request_id", handleDeleteEmailUpdateRequestRequest)
	router.Handle("POST", "/verify-new-email", handleUpdateEmailRequest)
	router.Handle("POST", "/users/:user_id/password-reset-requests", handleCreateUserPasswordResetRequestRequest)
	router.Handle("GET", "/password-reset-requests/:request_id", handleGetPasswordResetRequestRequest)
	router.Handle("DELETE", "/password-reset-requests/:request_id", handleDeletePasswordResetRequestRequest)
	router.Handle("POST", "/password-reset-requests/:request_id/verify-email", handleVerifyPasswordResetRequestEmailRequest)
	router.Handle("GET", "/users/:user_id/password-reset-requests", handleGetUserPasswordResetRequestsRequest)
	router.Handle("DELETE", "/users/:user_id/password-reset-requests", handleDeleteUserPasswordResetRequestsRequest)
	router.Handle("POST", "/reset-password", handleResetPasswordRequest)

	return router.Handler()

}
