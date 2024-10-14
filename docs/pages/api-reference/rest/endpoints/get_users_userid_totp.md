---
title: "GET /users/[user_id]/totp"
---

# GET /users/[user_id]/totp

Gets a user's TOTP credential.

```
GET https://your-domain.com/users/USER_ID/totp
```

## Response body

Returns the [TOTP credential model](/api-reference/rest/models/totp-credential) of the credential if it exists.

## Error codes

- [404] `NOT_FOUND`: The user or credential does not exist.
- [500] `UNKNOWN_ERROR`
