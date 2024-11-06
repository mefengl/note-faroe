---
title: "DELETE /totp-credentials/[credential_id]"
---

# DELETE /totp-credentials/[credential_id]

Deletes a TOTP credential.

```
DELETE https://your-domain.com/totp-credentials/CREDENTIAL_ID
```

## Successful response

No response body (204).

## Error codes

- [404] `NOT_FOUND`: The credential does not exist.
- [500] `INTERNAL_ERROR`
