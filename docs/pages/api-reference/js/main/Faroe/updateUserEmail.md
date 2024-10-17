---
title: "Faroe.updateUserEmail()"
---

# Faroe.updateUserEmail()

Mapped to [POST /update-email](/api-reference/rest/endpoints/post_update-email).

Updates a user's email with an email update request and returns the new user email. Upon a successful verification, all email update requests linked to the email address and password reset requests to the user are invalidated.

The update request is immediately invalidated after the 5th failed attempt.

## Definition

```ts
function updateUserEmail(requestId: string, code: string, clientIP: string): Promise<string>;
```

## Parameters

- `requestId`: A valid email update request ID.
- `code`: The verification code of the request.
- `clientIP`

## Error codes

- `TOO_MANY_REQUESTS`: Rate limit exceeded.
- `INCORRECT_CODE`: Incorrect verification code.
- `INVALID_REQUEST`: Invalid update request ID.
- `UNKNOWN_ERROR`
