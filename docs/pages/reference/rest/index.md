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

-   [User](/reference/rest/models/user)
-   [User email verification request](/reference/rest/models/user-email-verification-request)
-   [TOTP credential](/reference/rest/models/totp-credential)
-   [Password reset request](/reference/rest/models/password-reset-request)

## Endpoints

### Authentication

-   [POST /authenticate/password](/reference/rest/endpoints/post_authenticate_password): Authenticate user with email and password.

### Users

-   [POST /users](/reference/rest/endpoints/post_users): Create a new user.
-   [GET /users](/reference/rest/endpoints/get_users): Get a list of users.
-   [GET /users/\[user_id\]](/reference/rest/endpoints/get_users_userid): Get a user.
-   [DELETE /users/\[user_id\]](/reference/rest/endpoints/delete_users_userid): Delete a user.
-   [POST /users/\[user_id\]/update-password](/reference/rest/endpoints/post_users_userid_update-password): Update a user's password.

### Two-factor authentication

-   [POST /users/\[user_id\]/verify-recovery-code](/reference/rest/endpoints/post_users_userid_verify-recovery-code): Verify a user's recovery code.
-   [POST /users/\[user_id\]/regenerate-recovery-code](/reference/rest/endpoints/post_users_userid_regenerate-recovery-code): Generate a new user recovery code.
-   [DELETE /users/\[user_id\]/second-factors](/reference/rest/endpoints/delete_users_userid_second-factors): Deletes user's TOTP credentials.

### TOTP

-   [POST /users/\[user_id\/register-totp-credential](/reference/rest/endpoints/post_users_userid_register-totp-credential): Register a TOTP credential.
-   [GET /users/\[user_id\/totp-credentials](/reference/rest/endpoints/get_users_userid_totp-credentials): Get a list of a user's TOTP credential.
-   [DELETE /users/\[user_id\/totp-credentials](/reference/rest/endpoints/delete_users_userid_totp-credentials): Delete a user's TOTP credential.
-   [GET /totp-credentials/\[credential_id\]](/reference/rest/endpoints/get_totp-credentials-crendentialid): Get a TOTP credential.
-   [DELETE /totp-credentials/\[credential_id\]](/reference/rest/endpoints/delete_totp-credentials-crendentialid): Delete a TOTP credential.
-   [POST /totp-credentials/\[credential_id\]/verify-totp](/reference/rest/endpoints/post_totp-credentials-crendentialid_verify-totp): Verify a TOTP.

### Email verification

-   [POST /users/\[user_id\]/email-verification-request](/reference/rest/endpoints/post_users_userid_email-verification-request): Create a new user email verification request.
-   [GET /users/\[user_id\]/email-verification-request](/reference/rest/endpoints/get_users_userid_email-verification-request): Get a user's email verification request.
-   [DELETE /users/\[user_id\]/email-verification-request](/reference/rest/endpoints/delete_users_userid_email-verification-request): Delete a user's email verification request.
-   [POST /users/\[user_id\]/verify-email](/reference/rest/endpoints/post_users_userid_verify-email): Verify their email verification request code.

### Email update

-   [POST /users/\[user_id\]/email-update-requests](/reference/rest/endpoints/post_users_userid_email-update-requests): Create a new user email update request.
-   [GET /users/\[user_id\]/email-update-requests](/reference/rest/endpoints/get_users_userid_email-update-requests): Gets a list of a user's email update requests.
-   [DELETE /users/\[user_id\]/email-update-requests](/reference/rest/endpoints/delete_users_userid_email-update-requests): Deletes a user's email update requests.
-   [GET /email-update-requests/\[request_id\]](/reference/rest/endpoints/get_email-update-requests_requestid): Get an email update request.
-   [DELETE /email-update-requests/\[request_id\]](/reference/rest/endpoints/delete_email-update-requests_requestid): Delete an email update request.
-   [POST /verify-new-email](/reference/rest/endpoints/post_verify-new-email): Update a user's email by verifying their email update request code.

### Password reset

-   [POST /users/\[user_id\]/password-reset-requests](/reference/rest/endpoints/post_users_userid_password-reset-requests): Create a new password reset request for a user.
-   [GET /users/\[user_id\]/password-reset-requests](/reference/rest/endpoints/get_users_userid_password-reset-requests): Get a list of a user's password reset requests.
-   [DELETE /users/\[user_id\]/password-reset-requests](/reference/rest/endpoints/delete_users_userid_password-reset-requests): Delete a user's password reset requests.
-   [GET /password-reset-requests/\[request_id\]](/reference/rest/endpoints/get_password-reset-requests_requestid): Get a password reset request.
-   [DELETE /password-reset-requests/\[request_id\]](/reference/rest/endpoints/delete_password-reset-requests_requestid): Delete a password reset request.
-   [POST /password-reset-requests/\[request_id\]/verify-email](/reference/rest/endpoints/post_password-reset-requests_requestid_verify-email): Verify a reset request's email.
-   [POST /reset-password](/reference/rest/endpoints/post_reset-password): Reset the user's password with a verified reset request.
