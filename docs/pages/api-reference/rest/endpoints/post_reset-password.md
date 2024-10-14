---
title: "POST /reset-password"
---

# POST /reset-password

Resets a user's password with a password reset request. The password reset request must be marked as email-verified and, if the user has a second factor, it must be marked as 2fa-verfied.

On validation, it will mark the user's email as verified and invalidate all password reset requests linked to the user.

```
POST /reset-password
```

## Request body

All fields are required.

```ts
{
    "request_id": string,
    "password": string
}
```

- `request_id`: A valid password reset request ID.
- `password`: A new password. Must be at least 8 characters long and not be part of past data leaks (checked using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords)).

## Succesful response

Returns the [user model](/api-reference/rest/models/user) of the user. This endpoint can verify user emails.

## Error codess

- [400] `INVALID_DATA`: Invalid request data.
- [400] `PASSWORD_TOO_LARGE`: The password is too long.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [500] `UNKNOWN_ERROR`
