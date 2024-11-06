---
title: "DELETE /users/[user_id]/totp-credentials"
---

# DELETE /users/[user_id]/totp-credentials

Deletes all TOTP credentials of a user.

```
DELETE https://your-domain.com/users/[user_id]/totp-credentials
```

## Successful response

No response body (204).

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
