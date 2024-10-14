---
title: "Faroe.getPasswordResetRequest()"
---

# Faroe.getPasswordResetRequest()

Mapped to [GET /password-reset/\[request_id\]](/api-reference/rest/endpoints/get_password-reset_requestid).

Gets a password reset request. Returns `null` if the request doesn't exist.

## Definition

```ts
//$ FaroePasswordResetRequest=/api-reference/js/main/FaroePasswordResetRequest
async function getPasswordResetRequest(
    requestId: string
): Promise<$$FaroePasswordResetRequest | null>
```

### Parameters

- `requestId`

## Error codes

- `UNKNOWN_ERROR`
