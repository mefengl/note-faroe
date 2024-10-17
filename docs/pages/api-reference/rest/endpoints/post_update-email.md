---
title: "POST /update-email"
---

# POST /update-email

Updates a user's email with an email update request. Upon a successful verification, all email update requests linked to the email address and password reset requests to the user are invalidated.

The update request is immediately invalidated after the 5th failed attempt.

```
POST https://your-domain.com/update-email
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

The updated email.

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
