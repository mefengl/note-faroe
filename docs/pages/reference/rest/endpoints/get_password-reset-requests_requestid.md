---
title: "GET /password-reset-requests/[request_id]"
---

# GET /password-reset-requests/[request_id]

Gets a password reset request.

```
GET https://your-domain.com/password-reset-requests/REQUEST_ID
```

## Successful response

Returns the [password reset request model](/reference/rest/models/password-reset-request) if the request exists and is valid.

## Error codes

- [404] `NOT_FOUND`: The request does not exist or has expired.
- [500] `INTERNAL_ERROR`
