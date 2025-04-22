// Package main contains the core logic for the Faroe application, including HTTP handlers,
// database interactions, and background tasks. This file specifically handles
// operations related to user email verification and email updates.
package main

import (
	"context"      // Used for managing request lifecycles and cancellation signals.
	"database/sql" // Provides interfaces for interacting with SQL databases.
	"encoding/json" // Used for encoding and decoding JSON data.
	"errors"       // Provides functions for working with errors, like error checking.
	"fmt"           // Implements formatted I/O functions.
	"io"            // Provides basic I/O interfaces, used here for reading request bodies.
	"log"           // Used for logging messages, typically errors or informational notes.
	"net/http"      // Provides HTTP client and server implementations.
	"strings"       // Provides functions for string manipulation.
	"time"          // Provides functionality for measuring and displaying time.

	"github.com/julienschmidt/httprouter" // High-performance HTTP request router.
)

// handleCreateUserEmailVerificationRequestRequest handles API requests to initiate
// the email verification process for a given user. It generates a new verification
// code, stores it along with an expiration time, and sends back details about the
// request (excluding the code itself). Rate limiting is applied per user to prevent abuse.
//
// Security Checks:
// 1. Request Secret Verification: Ensures the request comes from a trusted client.
// 2. Accept Header Verification: Ensures the client accepts JSON responses.
// 3. User Existence Check: Verifies the target user ID exists.
// 4. Rate Limiting:
//    - Checks if the user has recently tried to verify (verifyUserEmailRateLimit).
//    - Consumes a token to limit how often verification requests can be *created* (createEmailRequestUserRateLimit).
//
// Parameters:
//   env (*Environment): Application environment containing database connections, secrets, rate limiters, etc.
//   w (http.ResponseWriter): The interface to write the HTTP response.
//   r (*http.Request): The incoming HTTP request details.
//   params (httprouter.Params): URL parameters extracted by the router (contains 'user_id').
func handleCreateUserEmailVerificationRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. Verify the shared secret included in the request headers.
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w) // 403 Forbidden if secret is invalid.
		return
	}
	// 2. Ensure the client accepts 'application/json' responses.
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w) // 406 Not Acceptable otherwise.
		return
	}

	// Extract the user ID from the URL path parameter.
	userId := params.ByName("user_id")
	// 3. Check if a user with this ID actually exists in the database.
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err) // Log unexpected database errors.
		writeUnexpectedErrorResponse(w) // 500 Internal Server Error.
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w) // 404 Not Found if the user doesn't exist.
		return
	}

	// 4. Apply Rate Limiting:
	// Check the rate limit for *verification attempts* for this user.
	// Although we are *creating* a request here, checking this prevents creating
	// new requests if the user is currently blocked due to too many failed *verification attempts*.
	if !env.verifyUserEmailRateLimit.Check(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests) // 429 Too Many Requests.
		return
	}
	// Consume a token from the rate limiter specific to *creating* verification requests.
	// This prevents a single user from spamming the creation endpoint.
	if !env.createEmailRequestUserRateLimit.Consume(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests) // 429 Too Many Requests.
		return
	}

	// Create the actual email verification request record in the database.
	// This generates a code and sets an expiration time.
	verificationRequest, err := createUserEmailVerificationRequest(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err) // Log errors during database insertion.
		// If creation failed, try to refund the rate limit token consumed earlier.
		env.createEmailRequestUserRateLimit.AddTokenIfEmpty(userId)
		writeUnexpectedErrorResponse(w) // 500 Internal Server Error.
		return
	}

	// Respond with the details of the created verification request (e.g., user ID, expiry).
	// Note: The actual verification code is NOT sent back in the response for security.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK.
	w.Write([]byte(verificationRequest.EncodeToJSON())) // Write JSON response body.
}

// handleVerifyUserEmailRequest handles API requests to verify a user's email address
// using a code provided by the user. It checks the code against the stored request,
// considering its expiration time and applying rate limits to prevent brute-force attacks.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. Content-Type Header Verification: Ensures the request body is JSON.
// 3. User Existence Check.
// 4. Verification Request Existence & Expiry Check.
// 5. Code Presence Check: Ensures a code was provided in the request body.
// 6. Rate Limiting: Consumes a token to limit verification *attempts* per user.
// 7. Code Validation: Compares the provided code with the stored code.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   params (httprouter.Params): URL parameters (contains 'user_id').
func handleVerifyUserEmailRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. Verify request secret.
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. Verify 'Content-Type' is 'application/json'.
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w) // 415 Unsupported Media Type.
		return
	}

	// 3. Check if user exists.
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

	// 4. Retrieve the existing email verification request for this user.
	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
	// If no request is found (ErrRecordNotFound)...
	if errors.Is(err, ErrRecordNotFound) {
		// Potentially refund a token for the *creation* rate limiter, allowing the user to try creating a new request.
		env.createEmailRequestUserRateLimit.AddTokenIfEmpty(userId)
		// Respond with 403 Not Allowed, indicating no active verification process to attempt.
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}
	// Handle other potential database errors.
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Check if the verification request has expired.
	// time.Now().Compare(t) returns:
	// -1 if time.Now() is before t
	//  0 if time.Now() is equal to t
	// +1 if time.Now() is after t
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 { // If expired (now is at or after ExpiresAt)
		// Attempt to delete the expired request from the database.
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			// Log deletion error but continue to respond as if it was just expired.
			log.Println(err)
		}
		// Refund the creation token and respond with 403 Not Allowed (expired).
		env.createEmailRequestUserRateLimit.AddTokenIfEmpty(userId)
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	// Read the JSON request body containing the verification code.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// If reading the body fails, it's likely invalid data.
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData) // 400 Bad Request.
		return
	}
	// Define a struct to unmarshal the JSON {"code": "..."}.
	var data struct {
		Code *string `json:"code"` // Pointer to handle potential null/missing field.
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		// JSON parsing failed.
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData) // 400 Bad Request.
		return
	}
	// 5. Check if the 'code' field was provided and is not empty.
	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData) // 400 Bad Request.
		return
	}

	// 6. Apply rate limiting for verification attempts.
	// Consume a token. If no tokens are available, the attempt is blocked.
	if !env.verifyUserEmailRateLimit.Consume(userId) {
		// If rate limited, delete the current verification request to force the user
		// to start a new verification process after the rate limit cooldown.
		// This prevents holding onto a potentially valid code while blocked.
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err) // Log deletion error.
			// Even if deletion fails, still respond with Too Many Requests.
		}
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests) // 429 Too Many Requests.
		return
	}

	// 7. Validate the provided code against the one stored in the database.
	// This function also typically deletes the request record upon successful validation.
	validCode, err := validateUserEmailVerificationRequest(env.db, r.Context(), userId, *data.Code)
	if err != nil {
		log.Println(err) // Log unexpected database errors during validation.
		writeUnexpectedErrorResponse(w) // 500 Internal Server Error.
		return
	}
	// If the code is incorrect...
	if !validCode {
		// Respond with 400 Bad Request (Incorrect Code).
		// Note: The rate limiter token was already consumed. Multiple incorrect attempts will lead to 429.
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	// If the code was valid and validation succeeded:
	// Reset the verification attempt rate limiter for this user, allowing them to
	// immediately start a new verification process if needed in the future.
	env.verifyUserEmailRateLimit.Reset(verificationRequest.UserId)

	// Respond with 204 No Content to indicate successful verification.
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteUserEmailVerificationRequestRequest handles API requests to explicitly
// delete an existing (non-expired) email verification request for a user. This might be
// used if the user wants to cancel the verification process.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. Verification Request Existence & Expiry Check.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   params (httprouter.Params): URL parameters (contains 'user_id').
func handleDeleteUserEmailVerificationRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. Verify request secret.
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	// Get user ID from URL.
	userId := params.ByName("user_id")
	// 2. Attempt to retrieve the verification request.
	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
	// If not found, respond with 404.
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	// Handle other potential database errors.
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Check if the request is already expired.
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		// If expired, attempt to delete it (cleanup).
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err) // Log deletion error but proceed.
		}
		// Respond with 404 Not Found, as the *active* request doesn't exist (it was expired).
		writeNotFoundErrorResponse(w)
		return
	}

	// If the request exists and is not expired, delete it.
	err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
	if err != nil {
		log.Println(err) // Log deletion error.
		writeUnexpectedErrorResponse(w) // Respond 500 if deletion fails.
		return
	}

	// Respond with 204 No Content on successful deletion.
	w.WriteHeader(http.StatusNoContent)
}

// handleGetUserEmailVerificationRequestRequest handles API requests to retrieve details
// about a pending email verification request for a user (e.g., its expiration time).
// It does NOT return the verification code itself.
//
// Security Checks:
// 1. Request Secret Verification.
// 2. Accept Header Verification: Ensures the client accepts JSON responses.
// 3. Verification Request Existence & Expiry Check.
//
// Parameters:
//   env (*Environment): Application environment.
//   w (http.ResponseWriter): HTTP response writer.
//   r (*http.Request): HTTP request.
//   params (httprouter.Params): URL parameters (contains 'user_id').
func handleGetUserEmailVerificationRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. Verify request secret.
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. Verify 'Accept' header is 'application/json'.
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	// Get user ID from URL.
	userId := params.ByName("user_id")
	// 3. Attempt to retrieve the verification request.
	verificationRequest, err := getUserEmailVerificationRequest(env.db, r.Context(), userId)
	// Handle not found error.
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	// Handle other database errors.
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// Check if the request is expired.
	if time.Now().Compare(verificationRequest.ExpiresAt) >= 0 {
		// If expired, attempt to delete it (cleanup).
		err = deleteUserEmailVerificationRequest(env.db, r.Context(), verificationRequest.UserId)
		if err != nil {
			log.Println(err) // Log deletion error but proceed.
		}
		// Respond with 404 Not Found, as the active request doesn't exist.
		writeNotFoundErrorResponse(w)
		return
	}

	// If found and not expired, respond with the request details (encoded as JSON).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK.
	w.Write([]byte(verificationRequest.EncodeToJSON()))
}

// getUserEmailVerificationRequest retrieves a pending email verification request
// from the database for a specific user ID.
//
// Parameters:
//   db (*sql.DB): Database connection pool.
//   ctx (context.Context): Request context for cancellation propagation.
//   userId (string): The ID of the user whose request is to be retrieved.
//
// Returns:
//   (UserEmailVerificationRequest): The found verification request details.
//   (error): ErrRecordNotFound if no request exists for the user, or any other
//            database error encountered during the query.
func getUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string) (UserEmailVerificationRequest, error) {
	// Retrieve the email verification request for the given user ID from the database.
	// This involves querying the 'user_email_verification_request' table.
	var verificationRequest UserEmailVerificationRequest
	// Variables to store Unix timestamps retrieved from the database.
	var createdAtUnix, expiresAtUnix int64
	// Query the database for the verification request row matching the user ID.
	row := db.QueryRowContext(ctx, "SELECT user_id, created_at, expires_at, code FROM user_email_verification_request WHERE user_id = ?", userId)
	// Scan the retrieved row columns into the verificationRequest struct fields and timestamp variables.
	err := row.Scan(&verificationRequest.UserId, &createdAtUnix, &expiresAtUnix, &verificationRequest.Code)
	// Check if the error is sql.ErrNoRows, indicating the record was not found.
	if errors.Is(err, sql.ErrNoRows) {
		// Return an empty request and the specific ErrRecordNotFound error.
		return UserEmailVerificationRequest{}, ErrRecordNotFound
	}
	// Convert the Unix timestamps (seconds since epoch) back into time.Time objects.
	verificationRequest.CreatedAt = time.Unix(createdAtUnix, 0)
	verificationRequest.ExpiresAt = time.Unix(expiresAtUnix, 0)
	// Return the populated request details and any potential scan error (other than ErrNoRows).
	return verificationRequest, err
}

// deleteUserEmailVerificationRequest deletes an email verification request
// from the database for a given user ID.
//
// Parameters:
//   db (*sql.DB): Database connection pool.
//   ctx (context.Context): Request context for cancellation propagation.
//   userId (string): The ID of the user whose request is to be deleted.
//
// Returns:
//   (error): Any database error encountered during the deletion.
func deleteUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string) error {
	// Delete the email verification request for the given user ID from the database.
	// This involves executing a DELETE query on the 'user_email_verification_request' table.
	_, err := db.ExecContext(ctx, "DELETE FROM user_email_verification_request WHERE user_id = ?", userId)
	// Return any error encountered during the deletion.
	return err
}

// validateUserEmailVerificationRequest attempts to redeem an email verification request
// by checking if the provided code matches the stored code for the user and if the
// request has not expired. If the code is valid and the request is not expired,
// the corresponding record is deleted from the database.
//
// Parameters:
//   db (*sql.DB): Database connection pool.
//   ctx (context.Context): Request context for cancellation propagation.
//   userId (string): The ID of the user attempting verification.
//   code (string): The verification code provided by the user.
//
// Returns:
//   (bool): True if the code was valid, the request was not expired, and the record
//           was successfully deleted. False otherwise.
//   (error): Any database error encountered during the deletion attempt.
func validateUserEmailVerificationRequest(db *sql.DB, ctx context.Context, userId string, code string) (bool, error) {
	// Execute a DELETE statement that targets the specific verification request row
	// matching the user ID, the provided code, and a non-expired timestamp.
	// The WHERE clause `expires_at > ?` ensures we only delete non-expired requests.
	result, err := db.ExecContext(ctx, "DELETE FROM user_email_verification_request WHERE user_id = ? AND code = ? AND expires_at > ?", userId, code, time.Now().Unix())
	if err != nil {
		// If there's a database error during execution, return false and the error.
		return false, err
	}
	// Check the number of rows affected by the DELETE statement.
	affected, err := result.RowsAffected()
	if err != nil {
		// If there's an error getting the affected rows count, return false and the error.
		return false, err
	}
	// If affected > 0, it means exactly one row was deleted, signifying that the
	// code was correct and the request was not expired.
	// If affected == 0, it means no matching, non-expired row was found (incorrect code or expired).
	return affected > 0, nil // Return true if a row was deleted, false otherwise, and nil error.
}

// UserEmailVerificationRequest defines the structure for storing user email verification data.
{{ ... }}
