---
title: "Faroe.resetPassword()"
---

# Faroe.resetPassword()

Mapped to [POST /reset-password](/api-reference/rest/endpoints/post_reset-password).

Resets a user's password with a password reset request. The password reset request must be marked as email-verified and, if the user has a second factor, it must be marked as 2FA-verfied.

On validation, it will mark the user's email as verified, invalidate all password reset requests linked to the user, and return the reset request's user.

*This method can verify user emails.*

## Definition

```ts
//$ FaroeUser=/api-reference/js/main/FaroeUser
async function resetPassword(requestId: string, password: string): Promise<$$FaroeUser>
```

### Parameters

- `request_id`: A valid password reset request ID.
- `password`: A new valid password. A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).

## Error codes

- `SECOND_FACTOR_NOT_VERIFIED`: 2FA required.
- `INVALID_DATA`: Invalid password length.
- `WEAK_PASSWORD`: The password is too weak.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `INVALID_REQUEST_ID`: Invalid reset request ID.
- `UNKNOWN_ERROR`
