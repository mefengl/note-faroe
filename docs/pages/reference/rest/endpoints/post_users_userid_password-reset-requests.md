---
title: "POST /users/[user_id]/password-reset-requests"
---

# POST /users/[user_id]/password-reset-requests

Creates a new password reset request for a user. This can only be called 3 times in a 15 minute window per user.

Send the created reset request's code to the email address.

```
POST https://your-domain.com/users/USER_ID/password-reset-requests
```

## Request body

```ts
{
    "client_ip": string
}
```

- `client_ip`: The client's IP address. If included, it will rate limit the endpoint based on it.

### Example

```json
{}
```

## Successful response

Returns the [password reset request model](/reference/rest/models/password-reset-requests-request) of the created request and a verification code. The code is only available here.

```ts
{
    "id": string,
    "user_id": string,
    "created_at": number,
    "expires_at": number,
    "code": string
}
```

- `code`: An 8-character alphanumeric email verification code.

### Example

```json
{
    "id": "cjjhw9ggvv7e9hfc3qjsiegv",
    "user_id": "wz2nyjz4ims4cyuw7eq6tnxy",
    "created_at": 1728804201,
    "expires_at": 1728804801,
    "twoFactorVerified": false
}
```

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
