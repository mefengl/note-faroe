---
title: "GET /users/[user_id]/totp-credentials"
---

# GET /users/[user_id]/totp-credentials

Gets a list of a user's TOTP credentials.

```
GET https://your-domain.com/users/USER_ID/totp-credentials
```

## Successful response

Returns a JSON array of [TOTP credential models](/reference/rest/models/totp-credential). If the user doesn't have any TOTP credentials, it will return an empty array.

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
