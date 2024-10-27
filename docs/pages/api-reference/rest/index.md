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

Faroe will return a 401 error response if the request has an invalid credential.

```json
{
    "error": "NOT_AUTHENTICATED"
}
```

## Responses

Successful responses will have a 200 status if it includes a response body or 204 status if not.

All error responses have a 4xx or 5xx status and includes a JSON object with an `error` field. See each endpoint's page for a list of possible response statuses and error codes.

```json
{
    "error": "INVALID_DATA"
}
```

## Data types

-   Email address: Must be less than 256 characters long, have a "@", and a "." in the domain part. Cannot start or end with a whitespace.
-   Password: Must be between 8 and 127 characters.

## Models

-   [User](/api-reference/rest/models/user)
-   [User email verification request](/api-reference/rest/models/user-email-verification-request)
-   [User TOTP credential](/api-reference/rest/models/user-totp-credential)
-   [Password reset request](/api-reference/rest/models/password-reset-request)

## Endpoints

### Authentication

-   [POST /authenticate/password](/api-reference/rest/endpoints/post_authenticate_password): Authenticate user with email and password.

### Users

-   [POST /users](/api-reference/rest/endpoints/post_users): Create a new user.
-   [GET /users](/api-reference/rest/endpoints/get_users): Get a list of users.
-   [GET /users/\[user_id\]](/api-reference/rest/endpoints/get_users_userid): Get a user.
-   [DELETE /users/\[user_id\]](/api-reference/rest/endpoints/delete_users_userid): Delete a user.
-   [POST /users/\[user_id\]/update-password](/api-reference/rest/endpoints/post_users_userid_update-password): Update a user's password.

#### Email verification

-   [POST /users/\[user_id\]/email-verification-request](/api-reference/rest/endpoints/post_users_userid_email-verification-request): Create a new user email verification request.
-   [GET /users/\[user_id\]/email-verification-request](/api-reference/rest/endpoints/get_users_userid_email-verification-request): Get a user's email verification request.
-   [DELETE /users/\[user_id\]/email-verification-request](/api-reference/rest/endpoints/delete_users_userid_email-verification-request): Delete a user's email verification request.
-   [POST /users/\[user_id\]/verify-email](/api-reference/rest/endpoints/post_users_userid_verify-email): Verify their email verification request code.

#### Email update

-   [POST /users/\[user_id\]/email-update-requests](/api-reference/rest/endpoints/post_users_userid_email-update-requests): Create a new user email update request.
-   [GET /users/\[user_id\]/email-update-requests](/api-reference/rest/endpoints/get_users_userid_email-update-requests): Gets a list of a user's email update requests.
-   [DELETE /users/\[user_id\]/email-update-requests](/api-reference/rest/endpoints/delete_users_userid_email-update-requests): Deletes a user's email update requests.
-   [GET /email-update-requests/\[request_id\]](/api-reference/rest/endpoints/get_email-update-requests_requestid): Get an email update request.
-   [DELETE /email-update-requests/\[request_id\]](/api-reference/rest/endpoints/delete_email-update-requests_requestid): Delete an email update request.
-   [POST /verify-new-email](/api-reference/rest/endpoints/post_verify-new-email): Update a user's email by verifying their email update request code.

#### Two-factor authentication

-   [POST /users/\[user_id\/register-totp](/api-reference/rest/endpoints/post_users_userid_register-totp): Register a TOTP credential.
-   [GET /users/\[user_id\]/totp-credential](/api-reference/rest/endpoints/get_users_userid_totp-credential): Get a user's TOTP credential.
-   [DELETE /users/\[user_id\]/totp-credential](/api-reference/rest/endpoints/delete_users_userid_totp-credential): Delete a user's TOTP credential.
-   [POST /users/\[user_id\]/verify-2fa/totp](/api-reference/rest/endpoints/post_users_userid_verify-2fa_totp): Verify a user's TOTP code.
-   [POST /users/\[user_id\]/regenerate-recovery-code](/api-reference/rest/endpoints/post_users_userid_regenerate-recovery-code): Generate a new user recovery code.
-   [POST /users/\[user_id\]/reset-2fa](/api-reference/rest/endpoints/post_users_userid_reset-2fa): Reset a user's second factors with a recovery code.

### Password reset

-   [POST /users/\[user_id\]/password-reset-requests](/api-reference/rest/endpoints/post_users_userid_password-reset-requests): Create a new password reset request for a user.
-   [GET /password-reset-requests/\[request_id\]](/api-reference/rest/endpoints/get_password-reset-requests_requestid): Get a password reset request.
-   [DELETE /password-reset-requests/\[request_id\]](/api-reference/rest/endpoints/delete_password-reset-requests_requestid): Delete a password reset request.
-   [POST /password-reset-requests/\[request_id\]/verify-email](/api-reference/rest/endpoints/post_password-reset-requests_requestid_verify-email): Verify a reset request's email.
-   [POST /reset-password](/api-reference/rest/endpoints/post_reset-password): Reset the user's password with a verified reset request.
