---
title: "GET /users/[user_id]/email-verification-request"
---

# GET /users/[user_id]/email-verification-request

Gets a user's valid email verification request.

```
GET https://your-domain.com/users/USER_ID/email-verification-reqes
```

## Successful response

Returns the [user email verification request model](/reference/rest/models/user-email-verification-request) if the request exists and is valid.

## Error codes

- [404] `NOT_FOUND`: The user doesn't exist, the user doesn't have a verification request, or their verification request has expired.
- [500] `INTERNAL_ERROR`
