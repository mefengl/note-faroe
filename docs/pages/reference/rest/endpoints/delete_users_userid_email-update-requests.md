---
title: "DELETE /users/[user_id]/email-update-request"
---

# DELETE /users/[user_id]/email-update-request

Deletes a user's email update requests.

```
DELETE https://your-domain.com/users/USER_ID/email-update-request"
```

## Successful response

No response body (204).

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
