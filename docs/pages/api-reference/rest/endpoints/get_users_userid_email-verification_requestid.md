---
title: "GET /users/[user_id]/email-verification/[request_id]"
---

# GET /users/[user_id]/email-verification/[request_id]

Gets a user's email verification request.

```
GET https://your-domain.com/users/USER_ID/email-verification/REQUEST_ID
```

## Succesful response

Returns the [email verification request model](/api-reference/rest/models/email-verification-request) if the request exists and is valid.

## Error codess

- [404] `NOT_FOUND`: The user or request does not exist, or the request has expired.
- [500] `UNKNOWN_ERROR`
