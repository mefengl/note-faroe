---
title: "POST /authenticate/password"
---

# POST /authenticate/password

Authenticates a user with email and password. It will temporary block the IP address if the client sends an incorrect password 5 times in a 15 minute window.

```
POST https://your-domain.com/authenticate/password
```

## Request body

All fields are required.

```ts
{
    "email": string,
    "password": string
}
```

- `email`
- `password`

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

- [400] `INVALID_DATA`: Invalid request data.
- [400] `USER_NOT_EXISTS`: The user does not exist.
- [400] `INCORRECT_PASSWORD`
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [500] `UNKNOWN_ERROR`
