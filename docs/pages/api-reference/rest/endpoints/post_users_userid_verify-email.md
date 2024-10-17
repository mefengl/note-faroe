---
title: "POST /users/[user_id]/verify-email"
---

# POST /users/[user_id]/verify-email

Verifies and updates a user's email with the user's email verification request code. The user will be locked out from verifying their email for 15 minutes after their 5th consecutive failed attempts.

```
POST https://your-domain.com/users/USER_ID/verify-email
```

## Request body

All fields are required.

```ts
{
    "code": string
}
```

- `code`: The verification code of the user's email verification request.

### Example

```json
{
    "code": "9TW45AZU"
}
```

## Successful response

No response body (204).

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `NOT_ALLOWED`: The user does not have a request or has a expired request.
- [400] `INCORRECT_CODE`: The one-time code is incorrect.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
