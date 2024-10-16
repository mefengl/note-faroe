---
title: "GET /email-verification-reqests/[request_id]"
---

# GET /users/[user_id]/email-verification-requests/[request_id]

Gets an email verification request.

```
GET https://your-domain.com/email-verification-requests/REQUEST_ID
```

## Succesful response

Returns the [email verification request model](/api-reference/rest/models/email-verification-request) if the request exists and is valid.

## Error codess

- [404] `NOT_FOUND`: The request does not exist or has expired.
- [500] `UNKNOWN_ERROR`
