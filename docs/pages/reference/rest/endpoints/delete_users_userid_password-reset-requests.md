---
title: "DELETE /users/[user_id]/password-reset-request"
---

# DELETE /users/[user_id]/password-reset-request

Deletes a user's password reset requests.

```
DELETE https://your-domain.com/users/USER_ID/password-reset-request"
```

## Successful response

No response body (204).

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
