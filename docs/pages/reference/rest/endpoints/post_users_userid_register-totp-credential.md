---
title: "POST /users/[user_id]/register-totp-credential"
---

# POST /users/[user_id]/register-totp-credential

Creates a TOTP credential. 

```
POST https://your-domain.com/users/USER_ID/register-totp-credential
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

## Successful response

Returns the [TOTP credential model](/reference/rest/models/totp-credential) of the registered credential.

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `INCORRECT_CODE`: Incorrect TOTP code.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
