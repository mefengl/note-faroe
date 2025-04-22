// Package main defines the entry point and core logic for the Faroe authentication server.
package main

import (
	"encoding/json" // Provides functionality for encoding and decoding JSON data.
	"errors"        // Provides functions to manipulate errors. Used here for checking specific error types (ErrRecordNotFound).
	"faroe/argon2id" // Custom package likely containing Argon2id password hashing functions (Verify).
	"io"            // Provides basic I/O primitives. Used here for reading the request body.
	"log"           // Provides simple logging capabilities. Used for logging unexpected errors.
	"net/http"      // Provides HTTP client and server implementations.

	"github.com/julienschmidt/httprouter" // High-performance HTTP request router.
)

// handleVerifyUserPasswordRequest handles requests to verify a user's password.
// It's likely used as part of a login flow or other actions requiring password confirmation.
//
// Security Checks Performed:
// 1. Request Secret Verification: Ensures the request comes from a trusted source (e.g., the frontend)
//    using a shared secret passed via a header or parameter (implementation detail in verifyRequestSecret).
// 2. Content-Type Verification: Checks if the request body is `application/json`.
// 3. Accept Header Verification: Checks if the client accepts `application/json` responses.
// 4. User Existence Check: Verifies that the user ID from the URL parameter corresponds to an existing user.
// 5. Rate Limiting: Applies rate limiting based on the client's IP address for both password hashing attempts
//    and general login attempts to mitigate brute-force attacks.
// 6. Password Verification: Uses Argon2id to securely compare the provided password against the stored hash.
//
// Parameters:
//   env (*Environment): Pointer to the application's environment containing shared resources like the database connection and secret key.
//   w (http.ResponseWriter): Used to write the HTTP response back to the client.
//   r (*http.Request): Represents the incoming HTTP request.
//   params (httprouter.Params): Contains the URL parameters extracted by the router (specifically, the 'user_id').
func handleVerifyUserPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. Verify the request secret to ensure the request originates from a trusted client.
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w) // Respond with 401 Not Authenticated if secret is invalid.
		return
	}
	// 2. Verify that the request body is JSON.
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w) // Respond with 415 Unsupported Media Type if Content-Type is not application/json.
		return
	}
	// 3. Verify that the client accepts JSON responses.
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w) // Respond with 406 Not Acceptable if Accept header doesn't include application/json.
		return
	}

	// Extract the user ID from the URL path parameters.
	userId := params.ByName("user_id")
	// Attempt to retrieve the user from the database using the extracted ID.
	user, err := getUser(env.db, r.Context(), userId)
	// 4. Handle potential errors during user retrieval.
	if errors.Is(err, ErrRecordNotFound) {
		// If the user is not found, respond with 404 Not Found.
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		// Log any other unexpected database errors and respond with 500 Internal Server Error.
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Read the entire request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Log errors during body reading and respond with 500.
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Define a struct to unmarshal the JSON request body.
	// Pointers are used for fields like Password to distinguish between a missing field and an empty string.
	var data struct {
		Password *string `json:"password"` // Pointer to the password string from the request.
		ClientIP string  `json:"client_ip"` // The client's IP address, provided in the request body (presumably by the frontend/proxy).
	}
	// Attempt to unmarshal the JSON body into the struct.
	err = json.Unmarshal(body, &data)
	if err != nil {
		// Log JSON parsing errors and respond with 400 Bad Request (Invalid Data).
		log.Println(err)
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// Validate that the password field was actually provided in the JSON.
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData) // Respond with 400 if password is missing.
		return
	}

	// 5. Apply Rate Limiting if ClientIP is provided.
	if data.ClientIP != "" {
		// Consume a token from the password hashing rate limiter for this IP.
		// This limits how often password *verification* can be attempted per IP.
		if !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
			writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests) // Respond with 429 Too Many Requests if limit exceeded.
			return
		}
		// Consume a token from the general login rate limiter for this IP.
		// This limits how often *any* login-related action can be attempted per IP.
		if !env.loginIPRateLimit.Consume(data.ClientIP) {
			writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests) // Respond with 429 if limit exceeded.
			return
		}
	}

	// 6. Verify the provided password against the stored hash using Argon2id.
	validPassword, err := argon2id.Verify(user.PasswordHash, *data.Password)
	if err != nil {
		// Log errors during password verification (should be rare) and respond with 500.
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Check if the password verification failed.
	if !validPassword {
		// Respond with a specific error for incorrect password (400 Bad Request).
		// Crucially, DO NOT reveal whether the user ID was valid or not here.
		// The rate limiting applied earlier helps mitigate guessing.
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}

	// If password verification was successful:
	if data.ClientIP != "" {
		// Replenish a token for the general login rate limiter if it was empty.
		// This might be used to slightly relax the limit after a successful login,
		// although consuming tokens on failure and adding only if empty on success seems unusual.
		// A more common pattern is simply resetting the failure count on success.
		env.loginIPRateLimit.AddTokenIfEmpty(data.ClientIP)
	}

	// Respond with 204 No Content upon successful password verification.
	// No response body is needed.
	w.WriteHeader(http.StatusNoContent) // Use http.StatusNoContent constant for clarity.
}
