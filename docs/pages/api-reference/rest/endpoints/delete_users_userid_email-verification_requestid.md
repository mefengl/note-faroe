---
title: "DELETE /users/[user_id]/email-verification/[request_id]"
---

# DELETE /users/[user_id]/email-verification/[request_id]

Deletes a user's email verification request.

```
DELETE https://your-domain.com/users/USER_ID/email-verification/REQUEST_ID
```

## Succesful response

No response body (204).

## Error codess

- [404] `NOT_FOUND`: The user or request does not exist, or the request has expired.
- [500] `UNKNOWN_ERROR`
