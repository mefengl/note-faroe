---
title: "POST /verify-new-email"
---

# POST /verify-new-email

Verifies an email update request's verification code. Upon a successful verification, all email update requests linked to the email address and password reset requests to the user are invalidated.

The update request is immediately invalidated after the 5th failed attempt.

```
POST https://your-domain.com/verify-new-email
```

## Request body

```ts
{
    "request_id": string,
    "code": string
}
```

- `request_id`: A valid email update request ID.
- `code`: The verification code of the request.

## Response body

The email address linked to the email update request.

```ts
{
    "email": string
}
```

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `TOO_MANY_REQUESTS`: Rate limit exceeded.
- [400] `INCORRECT_CODE`: Incorrect verification code.
- [400] `INVALID_REQUEST`: Invalid update request ID.
- [500] `UNKNOWN_ERROR`
