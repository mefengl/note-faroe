---
title: "Faroe.deletePasswordResetRequest()"
---

# Faroe.deletePasswordResetRequest()

Mapped to [DELETE /password-reset-requests/\[request_id\]](/reference/rest/endpoints/delete_password-reset-requests_requestid).

Deletes a passwod reset request. Deleting a non-existent request will not result in an error.

## Definition

```ts
async function deletePasswordResetRequest(requestId: string): Promise<void>
```

### Parameters

- `requestId`

## Error codes

- `INTERNAL_ERROR`
