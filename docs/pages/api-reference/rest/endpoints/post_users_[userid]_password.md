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

- `password`: Old password.
- `new_password`: Must be at least 8 characters long and not be part of past data leaks (checked using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords)).

### Example

```json
{
    "password": "48n2r3tnaqp",
    "new_password": "a83ri1lw2aw"
}
```

## Succesful response

Returns the [user model](/api-reference/rest/models/user) of the created user.

## Error codess

- [400] `INVALID_DATA`: Invalid request data.
- [400] `PASSWORD_TOO_LARGE`: The password is too long.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [500] `UNKNOWN_ERROR`
