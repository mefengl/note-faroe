---
title: "DELETE /users/[user_id]/totp"
---

# DELETE /users/[user_id]/totp

Deletes a user's TOTP credential.

```
DELETE https://your-domain.com/users/USER_ID/totp
```

## Successful response

No response body (204).

## Error codess

- [404] `NOT_FOUND`: The user does not exist or the user does not have a TOTP credential.
- [500] `UNKNOWN_ERROR`
