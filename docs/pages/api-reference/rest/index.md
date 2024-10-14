---
title: "REST API reference"
---

# REST API reference

All rest endpoints expects a JSON body and returns an JSON object.

## Authentication

Set the `Authorization` header to the credential you provided when initializing your server.

```
Authorization: YOUR_CREDENTIAL
```

## Responses

Successful responses will have a 200 status if it includes a response body or 204 status if not.

All error responses have a 4xx or 5xx status and includes a JSON object with an `error` field. See each endpoint's page for a list of possible response statuses and error codes.

```json
{
    "error": "INVALID_EMAIL"
}
```

## Rate limiting

Set the `X-Client-IP` header to the client's IP address to enable IP-based rate limiting.

```
X-Client-IP: 255.255.255.255
```

All endpoints that hashes passwords with Argon2id are rate limited based on IP addresses to prevent DoS attacks.

## Models

- [User](/api-reference/rest/models/user)
- [Email verification request](/api-reference/rest/models/email-verification-request)
- [TOTP credential](/api-reference/rest/models/totp-credential)
- [Password reset request](/api-reference/rest/models/password-reset-request)

## Endpoints

### Authentication

- [POST /login/password](/api-reference/rest/endpoints/post_authenticate_password): Authenticate user with email and password.

### Users

- [POST /users](/api-reference/rest/endpoints/post_users): Create a new user.
- [GET /users](/api-reference/rest/endpoints/get_users): Get a list of users.
- [GET /users/\[user_id\]](/api-reference/rest/endpoints/get_users_[userid]): Get a user.
- [DELETE /users/\[user_id\]](/api-reference/rest/endpoints/delete_users_[userid]): Delete a user.
- [POST /users/\[user_id\]/password](/api-reference/rest/endpoints/post_users_[userid]_password): Update a user's password.

#### Email verification

- [POST /users/\[user_id\]/email-verification](/api-reference/rest/endpoints/post_users_[userid]_email-verification): Create a new user email verification request.
- [GET /users/\[user_id\]/email-verification/\[request_id\]](/api-reference/rest/endpoints/get_users_[userid]_email-verification_[requestid]): Get a user's email verification request.
- [DELETE /users/\[user_id\]/email-verification/\[request_id\]](/api-reference/rest/endpoints/delete_users_[userid]_email-verification_[requestid]): Delete a user's email verification request.
- [POST /users/\[user_id\]/verify-email](/api-reference/rest/endpoints/post_users_[userid]_verify-email): Update the user's email by verifying their email verification request.

#### Two-factor authentication

- [POST /users/\[user_id\/totp](/api-reference/rest/endpoints/post_users_[userid]_totp): Register a TOTP credential.
- [GET /users/\[user_id\]/totp](/api-reference/rest/endpoints/get_users_[userid]_totp): Get a user's TOTP credential.
- [POST /users/\[user_id\]/verify-2fa/totp](/api-reference/rest/endpoints/post_users_[userid]_verify-2fa_totp): Verify a user's TOTP code.
- [GET /users/\[user_id\]/recovery-code](/api-reference/rest/endpoints/get_users_[userid]_recovery-code): Get a user's recovery code.
- [POST /users/\[user_id\]/regenerate-recovery-code](/api-reference/rest/endpoints/post_users_[userid]_regenerate-recovery-code): Generate a new user recovery code.
- [GET /users/\[user_id\]/reset-2fa](/api-reference/rest/endpoints/post_users_[userid]_reset-2fa): Reset a user's second factors with a recovery code.

### Password reset

- [POST /password-reset](/api-reference/rest/endpoints/post_password-reset): Create a new password reset request from the user's email.
- [GET /password-reset/\[request_id\]](/api-reference/rest/endpoints/get_password-reset_[requestid]): Get a password reset request.
- [DELETE /password-reset/\[request_id\]](/api-reference/rest/endpoints/delete_password-reset_[requestid]): Delete a password reset request.
- [POST /password-reset/\[request_id\]/verify-email](/api-reference/rest/endpoints/post_password-reset_[requestid]_verify-email): Verify a reset request's email.
- [POST /password-reset/\[request_id\]/verify-2fa/totp](/api-reference/rest/endpoints/post_password-reset_[requestid]_verify-2fa_totp): Verify the TOTP code of a reset request's user.
- [POST /password-reset/\[request_id\]/reset-2fa](/api-reference/rest/endpoints/post_password-reset_[requestid]_reset-2fa): Reset the second factors of a reset request's user with a recovery code.
- [POST /reset-password](/api-reference/rest/endpoints/post_reset-password): Reset the user's password with a verified reset request.
