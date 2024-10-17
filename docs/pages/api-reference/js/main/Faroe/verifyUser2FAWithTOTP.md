---
title: "Faroe.verifyUser2FAWithTOTP()"
---

# Faroe.verifyUser2FAWithTOTP()

Mapped to [POST /users/\[user_id\]/verify-2fa/totp](/api-reference/rest/endpoints/post_users_userid_verify-2fa_totp).

Verifies a user's TOTP code. The user will be locked out from using TOTP as their second factor for 15 minutes after their 5th consecutive failed attempts.

## Definition

```ts
async function verifyUser2FAWithTOTP(
    code: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `code`: The TOTP code.
- `clientIP`

## Error codes

- `NOT_ALLOWED`: The user does not have a TOTP credential registered.
- `TOO_MANY_REQUESTS`: Rate limit exceeded.
- `INCORRECT_CODE`: Incorrect TOTP code.
- `NOT_FOUND`: The user does not exist.
- `UNKNOWN_ERROR`
