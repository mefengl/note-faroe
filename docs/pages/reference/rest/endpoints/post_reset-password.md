---
title: "POST /reset-password"
---

# POST /reset-password

Resets a user's password with a password reset request. On validation, it will mark the user's email as verified and invalidate all password reset requests linked to the user.

```
POST /reset-password
```

## Request body

```ts
{
    "request_id": string,
    "password": string,
    "client_ip": string
}
```

- `request_id` (required): A valid password reset request ID.
- `password` (required): A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).
- `client_ip`: The client's IP address. If included, it will rate limit the endpoint based on it.

## Successful response

No response body (204).

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [400] `INVALID_REQUEST`: Invalid reset request ID.
- [500] `UNKNOWN_ERROR`
