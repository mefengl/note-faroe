---
title: "GET /email-update-requests/[request_id]"
---

# GET /email-update-requests/[request_id]

Gets an email update request.

```
GET https://your-domain.com/email-update-requests/[request_id]
```

## Successful response

Returns the [email update request model](/api-reference/rest/models/email-update-request) if the request exists and is valid.

## Error codes

- [404] `NOT_FOUND`: The request does not exist or has expired.
- [500] `UNKNOWN_ERROR`
