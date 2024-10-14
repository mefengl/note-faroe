---
title: "POST /users/[user_id]/reset-2fa"
---

# POST /users/[user_id]/reset-2fa

Resets a user's second factors using a recovery code and generates a new recovery code. The user will be locked out from using their recovery code for 15 minutes after their 5th consecutive failed attempts.

```
POST https://your-domain.com/users/USER_ID/reset-2fa
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
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
