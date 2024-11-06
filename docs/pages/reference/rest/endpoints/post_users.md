---
title: "POST /users"
---

# POST /users

Creates a new user.

We highly recommend putting a Captcha or equivalent in front for spam and bot detection.

```
POST https://your-domain.com/users
```

## Request body

```ts
{
    "password": string,
    "client_ip": string
}
```

- `password` (required): A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).
- `client_ip`: The client's IP address. If included, it will rate limit the endpoint based on it.

### Example

```json
{
    "password": "48n2r3tnaqp",
    "client_ip": "0.0.0.0"
}
```

## Successful response

Returns the [user model](/reference/rest/models/user) of the created user.

## Error codes

- [400] `INVALID_DATA`: Malformed email address; invalid password length.
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [500] `INTERNAL_ERROR`
