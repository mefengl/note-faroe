---
title: "DELETE /email-verification-requests/[request_id]"
---

# DELETE /email-verification-requests/[request_id]

Deletes an email verification request.

```
DELETE https://your-domain.com/email-verification-requests/REQUEST_ID
```

## Succesful response

No response body (204).

## Error codess

- [404] `NOT_FOUND`: The request does not exist or has expired.
- [500] `UNKNOWN_ERROR`
