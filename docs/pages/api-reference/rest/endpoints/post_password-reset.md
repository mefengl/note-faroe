---
title: "POST /password-reset"
---

# POST /password-reset

Creates a new password reset request for a user. This can only be called 3 times in a 15 minute window per user.

Send the created reset request's code to the email address.

```
POST https://your-domain.com/password-reset
```

## Request body

All fields are required.

```ts
{
    "email": string
}
```

- `email`

### Example

```json
{
    "email": "penguin@example.com"
}
```

## Succesful response

Returns the [password reset request model](/api-reference/rest/models/password-reset-request) of the created request and a verification code. The code is only available here.

```ts
{
    "id": string,
    "user_id": string,
    "created_at": number,
    "expires_at": number,
    "email": string,
    "email_verified": boolean,
    "twoFactorVerified": boolean,
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
    "email": "cat@example.com",
    "email_verified": true,
    "twoFactorVerified": false
}
```

## Error codess

- [400] `INVALID_DATA`: Invalid request data.
- [400] `USER_NOT_EXISTS`: A user linked to the email does not exist.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [500] `UNKNOWN_ERROR`
