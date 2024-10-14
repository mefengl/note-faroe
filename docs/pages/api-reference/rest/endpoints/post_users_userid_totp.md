---
title: "POST /users/[user_id]/totp"
---

# POST /users/[user_id]/totp

Registers a TOTP (SHA-1, 6 digits, 30 seconds interval) credential to a user.

```
POST https://your-domain.com/users/USER_ID/totp
```

## Request body

All fields are required.

```ts
{
    "totp_key": string,
    "code": string
}
```

- `totp_key`: A base64-encoded TOTP key. The encoded key must be 20 bytes.
- `code`: The TOTP code from the key for verification.

## Response body

Returns the [TOTP credential model](/api-reference/rest/models/totp-credential) of the registered credential.

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `INCORRECT_CODE`: Incorrect TOTP code.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
