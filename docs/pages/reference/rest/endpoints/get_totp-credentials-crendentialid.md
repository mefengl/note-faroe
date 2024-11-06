---
title: "GET /totp-credentials/[credential_id]"
---

# GET /totp-credentials/[credential_id]

Gets a TOTP credential.

```
GET https://your-domain.com/totp-credentials/CREDENTIAL_ID
```

## Successful response

Returns the [TOTP credential model](/reference/rest/models/password-reset-request) if the credential exists.

## Error codes

-   [404] `NOT_FOUND`: The request does not exist or has expired.
-   [500] `INTERNAL_ERROR`
