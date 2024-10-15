---
title: "POST /reset-password"
---

# POST /reset-password

Resets a user's password with a password reset request. The password reset request must be marked as email-verified.

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
- `password`: A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).

## Succesful response

No response body (204).

## Error codess

- [400] `INVALID_DATA`: Invalid request data.
- [400] `EMAIL_NOT_VERIFIED`: Reset request email not verified.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [400] `INVALID_REQUEST_ID`: Invalid reset request ID.
- [500] `UNKNOWN_ERROR`
