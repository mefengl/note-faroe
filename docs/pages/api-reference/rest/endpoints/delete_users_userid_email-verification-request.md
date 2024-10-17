---
title: "DELETE /users/[user_id]/email-verification-request"
---

# DELETE /users/[user_id]/email-verification-request

Deletes a user's email verification request.

```
DELETE https://your-domain.com/users/USER_ID/email-verification-request"
```

## Succesful response

No response body (204).

## Error codess

- [404] `NOT_FOUND`: The request or user does not exist, or the request has expired.
- [500] `UNKNOWN_ERROR`
