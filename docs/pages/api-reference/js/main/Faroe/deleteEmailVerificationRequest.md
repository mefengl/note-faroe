---
title: "Faroe.deleteEmailVerificationRequest()"
---

# Faroe.deleteEmailVerificationRequest()

Mapped to [DELETE /email-verification-requests/\[request_id\]](/api-reference/rest/endpoints/delete_email-verification-requests_requestid).

Deletes a user's email verification request. Deleting a non-existent request will not result in an error.

## Definition

```ts
async function deleteEmailVerificationRequest(requestId: string): Promise<void>
```

### Parameters

- `userId`
- `requestId`

## Error codes

- `UNKNOWN_ERROR`
