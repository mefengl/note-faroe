---
title: "DELETE /users/[user_id]/email-verification-request"
---

# DELETE /users/[user_id]/email-verification-request

Deletes a user's email verification request.

```
DELETE https://your-domain.com/users/USER_ID/email-verification-request"
```

## Successful response

No response body (204).

## Error codes

- [404] `NOT_FOUND`: The user doesn't exist, the user doesn't have a verification request, or their verification request has expired.
- [500] `INTERNAL_ERROR`
