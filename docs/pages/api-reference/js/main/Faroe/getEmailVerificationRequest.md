---
title: "Faroe.getEmailVerificationRequest()"
---

# Faroe.getEmailVerificationRequest()

Mapped to [GET /email-verification-requests/\[request_id\]](/api-reference/rest/endpoints/get_email-verification-requests_requestid).

Gets a email verification request. Returns `null` if the request doesn't exist or has expired.

## Definition

```ts
//$ FaroeEmailVerificationRequest=/api-reference/js/main/FaroeEmailVerificationRequest
async function getEmailVerificationRequest(
    requestId: string
): Promise<$$FaroeEmailVerificationRequest | null>
```

### Parameters

- `userId`
- `requestId`

## Error codes

- `UNKNOWN_ERROR`
