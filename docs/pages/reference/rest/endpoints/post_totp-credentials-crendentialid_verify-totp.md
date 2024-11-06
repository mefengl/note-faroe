---
title: "POST /totp-credentials/[credential_id]/verify-totp"
---

# POST /totp-credentials/[credential_id]/verify-totp

Verifies a TOTP of a TOTP credential.

```
POST https://your-domain.com/totp-credentials/CREDENTIAL_ID/verify-totp
```

## Request body

All fields are required.

```json
{
    "code": string
}
```

- `code`: The TOTP code.

## Successful response

No response body (204).

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `INCORRECT_CODE`: Incorrect TOTP code.
- [404] `NOT_FOUND`: The credential does not exist.
- [500] `INTERNAL_ERROR`
