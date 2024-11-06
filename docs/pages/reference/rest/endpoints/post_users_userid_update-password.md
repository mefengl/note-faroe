---
title: "POST /users/[user_id]/update-password"
---

# POST /users/[user_id]/update-password

Updates a user's password.

```
POST https://your-domain.com/users/USER_ID/update-password
```

## Request body

```ts
{
    "password": string,
    "new_password": string,
    "client_ip": string
}
```

- `password` (required): The current password.
- `new_password` (required): A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).
- `client_ip`: The client's IP address. If included, it will rate limit the endpoint based on it.

### Example

```json
{
    "password": "48n2r3tnaqp",
    "new_password": "a83ri1lw2aw",
    "client_ip": "0.0.0.0"
}
```

## Successful response

No response body (204).

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
