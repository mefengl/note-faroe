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

All fields are required.

```ts
{
    "email": string,
    "password": string
}
```

- `email`: A valid email address.
- `password`: A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).

### Example

```json
{
    "email": "penguin@example.com",
    "password": "48n2r3tnaqp"
}
```

## Succesful response

Returns the [user model](/api-reference/rest/models/user) of the created user.

## Error codess

- [400] `INVALID_DATA`: Malformed email address; invalid password length.
- [400] `EMAIL_ALREADY_USED`
- [400] `WEAK_PASSWORD`: The password is too weak.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [500] `UNKNOWN_ERROR`
