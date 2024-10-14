---
title: "POST /users/[user_id]/password"
---

# POST /users/[user_id]/password

Updates a user's password.

```
POST https://your-domain.com/users/USER_ID/password
```

## Request body

All fields are required.

```ts
{
    "password": string,
    "new_password": string
}
```

- `password`: The current password.
- `new_password`: A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).

### Example

```json
{
    "password": "48n2r3tnaqp",
    "new_password": "a83ri1lw2aw"
}
```

## Succesful response

No response body (204).

## Error codess

- [400] `INVALID_DATA`: Invalid request data.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
