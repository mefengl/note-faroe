---
title: "Faroe.verifyNewUserEmail()"
---

# Faroe.verifyNewUserEmail()

Mapped to [POST /verify-new-email](/api-reference/rest/endpoints/post_verify-new-email).

Verifies an email update request's verification code. Upon a successful verification, all email update requests linked to the email address and password reset requests to the user are invalidated.

The update request is immediately invalidated after the 5th failed attempt.

## Definition

```ts
function verifyNewUserEmail(requestId: string, code: string): Promise<string>;
```

## Parameters

- `requestId`: A valid email update request ID.
- `code`: The verification code of the request.

## Error codes

- `TOO_MANY_REQUESTS`: Rate limit exceeded.
- `INCORRECT_CODE`: Incorrect verification code.
- `INVALID_REQUEST`: Invalid update request ID.
- `EMAIL_ALREADY_USED`: Email is already used.
- `UNKNOWN_ERROR`
