---
title: "POST /password-reset/[request_id]/reset-2fa"
---

# POST /password-reset/[request_id]/reset-2fa

Resets the second factors of a password reset request's user using a recovery code and generates a new recovery code. The user will be locked out from using their recovery code for 15 minutes after their 5th consecutive failed attempts.

```
POST https://your-domain.com/password-reset/REQUEST_ID/reset-2fa
```

## Request body

All fields are required.

```ts
{
    "recovery_code": string
}
```

- `recovery_code`

## Successful response

Return the user's new recovery code.

```ts
{
    "recovery_code": string
}
```

### Example

```json
{
    "recovery_code": "4UHZRTWP"
}
```

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `TOO_MANY_REQUESTS`: Rate limit exceeded.
- [400] `INCORRECT_CODE`: Incorrect recovery code.
- [404] `NOT_FOUND`: The reset request does not exist.
- [500] `UNKNOWN_ERROR`
