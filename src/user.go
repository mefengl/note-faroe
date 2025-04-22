// Package main defines the entry point and core logic for the Faroe authentication server.
package main

import (
	"bufio"         // Provides buffered I/O operations, used here for writing formatted user lists.
	"context"       // Manages deadlines, cancellation signals, and other request-scoped values across API boundaries.
	"crypto/sha1"   // Provides SHA1 hashing algorithm, used here for checking against the Pwned Passwords database.
	"database/sql"  // Provides a generic interface around SQL (or SQL-like) databases.
	"encoding/hex"  // Provides hex encoding and decoding.
	"encoding/json" // Provides functionality for encoding and decoding JSON data.
	"errors"        // Provides functions to manipulate errors.
	"faroe/argon2id" // Custom package likely containing Argon2id password hashing functions.
	"fmt"           // Provides functions for formatted I/O.
	"io"            // Provides basic I/O primitives.
	"log"           // Provides simple logging capabilities.
	"math"          // Provides basic mathematical constants and functions.
	"net/http"      // Provides HTTP client and server implementations.
	"regexp"        // Provides regular expression searching.
	"strconv"       // Provides conversions to and from string representations of basic data types.
	"strings"       // Provides functions for string manipulation.
	"time"          // Provides functionality for measuring and displaying time.

	"github.com/julienschmidt/httprouter" // High-performance HTTP request router.
)

// handleCreateUserRequest handles requests to create a new user account.
// It validates the provided password for strength, hashes it securely using Argon2id,
// applies rate limiting based on IP for hashing, and then inserts the new user into the database.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. Content-Type and Accept Header Verification (JSON).
// 3. Password Validation: Checks if the password is provided, not empty, and within length limits (<= 127 chars).
// 4. Password Strength Check: Verifies the password against common patterns and potentially a database of breached passwords (like Pwned Passwords via Have I Been Pwned API, though the check here seems simpler based on `verifyPasswordStrength` implementation).
// 5. Rate Limiting: Limits password hashing attempts per IP address.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   _ (httprouter.Params): URL parameters (not used in this handler).
func handleCreateUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Standard request verification (secret, content-type, accept).
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

	// Read request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Define struct for JSON request body.
	var data struct {
		Password *string `json:"password"` // User's chosen password.
		ClientIP string  `json:"client_ip"` // Client's IP for rate limiting.
	}
	// Unmarshal JSON data.
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// Validate password presence and basic constraints.
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if *data.Password == "" || len(*data.Password) > 127 { // Check for empty or overly long password.
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// Verify password strength.
	strongPassword, err := verifyPasswordStrength(*data.Password)
	if err != nil {
		log.Println(err) // Log errors during strength check.
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword) // Respond if password is weak.
		return
	}

	// Apply rate limiting before expensive hashing operation.
	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	// Hash the password using Argon2id.
	passwordHash, err := argon2id.Hash(*data.Password)
	if err != nil {
		log.Println(err) // Log errors during hashing.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Create the user record in the database.
	user, err := createUser(env.db, r.Context(), passwordHash)
	if err != nil {
		log.Println(err) // Log errors during database insertion.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Respond with the newly created user's details (encoded as JSON).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Use http.StatusOK for clarity.
	w.Write([]byte(user.EncodeToJSON()))
}

// handleGetUserRequest handles requests to retrieve details for a specific user.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. Accept Header Verification (JSON).
// 3. User Existence Check.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   params (httprouter.Params): URL parameters, containing 'user_id'.
func handleGetUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Standard request verification (secret, accept).
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	// Get user ID from URL parameters.
	userId := params.ByName("user_id")
	// Fetch user from the database.
	user, err := getUser(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w) // Respond 404 if user not found.
		return
	}
	if err != nil {
		log.Println(err) // Log other database errors.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Respond with the user's details (encoded as JSON).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Use http.StatusOK.
	w.Write([]byte(user.EncodeToJSON()))
}

// handleDeleteUserRequest handles requests to delete a specific user account.
// It first checks if the user exists before attempting deletion.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. User Existence Check.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   params (httprouter.Params): URL parameters, containing 'user_id'.
func handleDeleteUserRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Standard request verification (secret).
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	// Get user ID from URL parameters.
	userId := params.ByName("user_id")
	// Check if the user exists before trying to delete.
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err) // Log database errors during check.
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w) // Respond 404 if user doesn't exist.
		return
	}

	// Attempt to delete the user from the database.
	err = deleteUser(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err) // Log errors during deletion.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Respond with 204 No Content on successful deletion.
	w.WriteHeader(http.StatusNoContent) // Use http.StatusNoContent.
}

// handleUpdateUserPasswordRequest handles requests to update a user's password.
// It requires the current password for verification before updating to the new password.
// It performs strength checks on the new password and applies rate limiting.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. Content-Type Header Verification (JSON).
// 3. User Existence Check.
// 4. Current Password Verification (using Argon2id).
// 5. New Password Validation: Checks presence, constraints (not empty, <= 127 chars).
// 6. New Password Strength Check.
// 7. Rate Limiting: Limits password hashing attempts per IP.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   params (httprouter.Params): URL parameters, containing 'user_id'.
func handleUpdateUserPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Standard request verification (secret, content-type).
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	// Get user ID and fetch user data.
	userId := params.ByName("user_id")
	user, err := getUser(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w) // Respond 404 if user not found.
		return
	}
	if err != nil {
		log.Println(err) // Log other database errors.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Read request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Define struct for JSON request body.
	var data struct {
		Password    *string `json:"password"`     // Current password for verification.
		NewPassword *string `json:"new_password"` // The desired new password.
		ClientIP    string  `json:"client_ip"`    // Client's IP for rate limiting.
	}
	// Unmarshal JSON data.
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// Validate presence of current password.
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// Assign pointers to local variables for easier use (handle potential nil dereference if NewPassword is nil below).
	password := *data.Password // Note: Potential panic if data.Password was nil, but checked above.
	var newPassword string
	if data.NewPassword != nil { // Check if NewPassword was provided before dereferencing.
		newPassword = *data.NewPassword
	} else {
		// If NewPassword is nil (not provided in JSON), treat it as an invalid request.
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// Validate password constraints.
	if password == "" || len(password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if newPassword == "" || len(newPassword) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// Verify the current password provided by the user against the stored hash.
	// This uses the argon2id.ComparePasswordAndHash function for secure comparison.
	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		log.Println(err) // Log errors during password comparison.
		writeUnexpectedErrorResponse(w)
		return
	}
	// If the current password doesn't match the stored hash, return an authentication error.
	if !match {
		writeExpectedErrorResponse(w, ExpectedErrorAuthenticationFailed)
		return
	}

	// Check the strength of the new password using the verifyPasswordStrength function.
	// This helps prevent users from choosing weak or easily guessable passwords.
	strongPassword, err := verifyPasswordStrength(newPassword)
	if err != nil {
		log.Println(err) // Log errors during strength check.
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorPasswordTooWeak)
		return
	}

	// Apply rate limiting before hashing the new password.
	// This uses the client's IP address to limit the number of password hashing attempts
	// from a single source, mitigating brute-force or resource exhaustion attacks.
	if !env.rateLimiter.Allow(data.ClientIP) {
		writeTooManyRequestsErrorResponse(w)
		return
	}

	// Hash the new password using Argon2id before storing it.
	// Argon2id is a secure, memory-hard hashing algorithm recommended for password storage.
	newPasswordHash, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
	if err != nil {
		log.Println(err) // Log errors during hashing.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Update the user's password hash in the database with the new hash.
	err = updateUserPassword(env.db, r.Context(), userId, newPasswordHash)
	if err != nil {
		log.Println(err) // Log errors during the database update.
		writeUnexpectedErrorResponse(w)
		return
	}

	// Respond with 204 No Content to indicate successful password update.
	w.WriteHeader(http.StatusNoContent)
}
