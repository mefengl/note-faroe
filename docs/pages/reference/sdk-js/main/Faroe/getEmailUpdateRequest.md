---
title: "Faroe.getEmailUpdateRequest()"
---

# Faroe.getEmailUpdateRequest()

Mapped to [GET /email-update-requests/\[request_id\]](/reference/rest/endpoints/get_email-update-requests_requestid).

Gets a email update request. Returns `null` if the request doesn't exist or has expired.

## Definition

```ts
//$ FaroeEmailUpdateRequest=/reference/sdk-js/main/FaroeEmailUpdateRequest
async function getEmailUpdateRequest(
    requestId: string
): Promise<$$FaroeEmailUpdateRequest | null>
```

### Parameters

- `requestId`

## Error codes

- `INTERNAL_ERROR`
