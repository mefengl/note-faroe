---
title: "POST /password-reset-requests/[request_id]/verify-email"
---

# POST /password-reset-requests/[request_id]/verify-email

Verifies the email linked to a password reset request with a verification code.

The reset request is immediately invalidated after the 5th failed attempt.

```
POST https://your-domain.com/password-reset-requests/REQUEST_ID/verify-email
```

## Request body

```ts
{
    "code": string,
    "client_ip": string
}
```

- `code` (required): The email verification code for the password reset request.
- `client_ip`: The client's IP address. If included, it will rate limit the endpoint based on it.

### Example

```json
{
    "code": "9TW45AZU",
    "client_ip": "0.0.0.0"
}
```

## Successful response

No response body (204).

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `INCORRECT_CODE`: The one-time code is incorrect.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The password reset request does not exist or has expired.
- [500] `INTERNAL_ERROR`
