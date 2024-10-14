---
title: "Faroe.verifyPasswordResetRequest2FAWithTOTP()"
---

# Faroe.verifyPasswordResetRequest2FAWithTOTP()

Mapped to [POST /password-reset/\[request_id\]/verify-2fa/totp](/api-reference/rest/endpoints/post_password-reset_requestid_verify-2fa_totp).

Verifies the TOTP code of a password reset request's user and marks the password reset as 2FA-verfied. The user will be locked out from using TOTP as their second factor for 15 minutes after their 5th consecutive failed attempts.

## Definition

```ts
async function verifyPasswordResetRequest2FAWithTOTP(
    code: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `code`: TOTP code.
- `clientIP`

## Error codes

- `SECOND_FACTOR_NOT_ALLOWED`: The user does not have a TOTP credential registered.
- `TOO_MANY_REQUESTS`: Rate limit exceeded.
- `INCORRECT_CODE`: Incorrect TOTP code.
- `NOT_FOUND`: The reset request does not exist.
- `UNKNOWN_ERROR`
