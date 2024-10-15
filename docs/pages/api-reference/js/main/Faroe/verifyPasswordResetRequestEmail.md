---
title: "Faroe.verifyPasswordResetRequestEmail()"
---

# Faroe.verifyPasswordResetRequestEmail()

Mapped to

Verifies the email linked to a password reset request with a verification code.

The reset request is immediately invalidated after the 5th failed attempt.

## Definition

```ts
async function verifyPasswordResetRequestEmail(
    requestId: string,
    code: string,
    clientIP: string
): Promise<void>
```

### Parameters

- `requestId`
- `code`
- `clientIP`

## Error codes

- `INCORRECT_CODE`: The one-time code is incorrect.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The password reset request does not exist or has expired.
- `UNKNOWN_ERROR`
