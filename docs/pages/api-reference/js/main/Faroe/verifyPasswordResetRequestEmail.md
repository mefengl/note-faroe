---
title: "Faroe.verifyPasswordResetRequestEmail()"
---

# Faroe.verifyPasswordResetRequestEmail()

Mapped to [POST /password-reset/\[request_id\]/verify-email](/api-reference/rest/endpoints/post_password-reset_requestid_verify-email).

Verifies the email linked to a password reset request with a verification code.

The reset request is immediately invalidated after the 5th failed attempt.

## Definition

```ts
async function verifyPasswordResetRequestEmail(
    code: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `code`: The email verification code for the password reset request.
- `clientIP`

## Error codes

- [400] `INCORRECT_CODE`: The one-time code is incorrect.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The password reset request does not exist or has expired.
- [500] `UNKNOWN_ERROR`
