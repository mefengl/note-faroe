---
title: "GET /password-reset/[request_id]"
---

# GET /password-reset/[request_id]

Gets a password reset request.

```
GET https://your-domain.com/password-reset/REQUEST_ID
```

## Succesful response

Returns the [password reset request model](/api-reference/rest/models/password-reset-request) if the request exists and is valid.

## Error codess

- [404] `NOT_FOUND`: The request does not exist or the request has expired.
- [500] `UNKNOWN_ERROR`
