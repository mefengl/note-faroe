---
title: "DELETE /users/[user_id]/second-factors"
---

# DELETE /users/[user_id]/second-factors

Deletes all TOTP credentials of user.

```
DELETE https://your-domain.com/users/USER_ID/second-factors
```

## Successful response

No response body (204).

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
