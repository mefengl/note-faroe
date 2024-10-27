---
title: "POST /users/[user_id]/verify-password"
---

# POST /users/[user_id]/verify-password

Verifies a user's password. It will temporary block the IP address if the client sends an incorrect password 5 times in a 15 minute window.

```
POST https://your-domain.com/users/USER_ID/verify-password
```

## Request body

```ts
{
    "password": string,
    "client_ip": string
}
```

- `password` (required): A valid password.
- `client_ip`: The client's IP address. If included, it will rate limit the endpoint based on it.

### Example

```json
{
    "password": "48n2r3tnaqp"
}
```

## Successful response

No response body (204).

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `INCORRECT_PASSWORD`
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
