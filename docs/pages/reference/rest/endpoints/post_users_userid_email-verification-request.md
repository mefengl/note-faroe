---
title: "POST /users/[user_id]/email-verification-request"
---

# POST /users/[user_id]/email-verification-request

Creates a new email verification request for a user. This can only be called 3 times in a 15 minute window per user.

```
POST https://your-domain.com/users/USER_ID/email-verification-request
```

## Successful response

Returns the [user email verification request model](/reference/rest/models/user-email-verification-request) of the created request.

## Error codes

- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
