---
title: "Faroe.resetPasswordResetRequestUser2FA()"
---

# Faroe.resetPasswordResetRequestUser2FA()

Mapped to [POST /password-reset/\[request_id\]/reset-2fa](/api-reference/rest/endpoints/post_password-reset_requestid_reset-2fa).

Resets the second factors of a password reset request's user using a recovery code and returns a new recovery code. The user will be locked out from using their recovery code for 15 minutes after their 5th consecutive failed attempts.

## Definition

```ts
async function resetPasswordResetRequestUser2FA(
    requestId: string,
    recoveryCode: string,
    clientIP: string | null
): Promise<string>
```

### Parameters

- `requestId`
- `recoveryCode`
- `clientIP`

## Error codes

- `TOO_MANY_REQUESTS`: Rate limit exceeded.
- `INCORRECT_CODE`: Incorrect recovery code.
- `NOT_FOUND`: The reset request does not exist.
- `UNKNOWN_ERROR`
