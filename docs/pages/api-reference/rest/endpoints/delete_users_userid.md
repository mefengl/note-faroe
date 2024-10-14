---
title: "DELETE /users/[user_id]"
---

# DELETE /users/[user_id]

Deletes a user.

```
DELETE https://your-domain.com/users/USER_ID
```

## Succesful response

No response body (204).

## Error codess

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
