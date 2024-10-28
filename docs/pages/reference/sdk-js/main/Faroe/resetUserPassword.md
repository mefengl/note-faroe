---
title: "Faroe.resetUserPassword()"
---

# Faroe.resetUserPassword()

Mapped to [POST /reset-password](/reference/rest/endpoints/post_reset-password).

Resets a user's password with a password reset request.On validation, it will mark the user's email as verified and invalidate all password reset requests linked to the user.

## Definition

```ts
async function resetUserPassword(
    requestId: string,
    password: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `request_id`: A valid password reset request ID.
- `password`: A new valid password. A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).
- `clientIP`

## Error codes

- `SECOND_FACTOR_NOT_VERIFIED`: 2FA required.
- `INVALID_DATA`: Invalid password length.
- `WEAK_PASSWORD`: The password is too weak.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `INVALID_REQUEST`: Invalid reset request ID.
- `UNKNOWN_ERROR`
