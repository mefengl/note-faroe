---
title: "DELETE /password-reset/[request_id]"
---

# DELETE /password-reset/[request_id]

Deletes a passwod reset request.

```
DELETE https://your-domain.com/password-reset/REQUEST_ID
```

## Succesful response

No response body (204).

## Error codess

- [404] `NOT_FOUND`: The request does not exist or has expired.
- [500] `UNKNOWN_ERROR`
