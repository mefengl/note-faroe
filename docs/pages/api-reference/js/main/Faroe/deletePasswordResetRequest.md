---
title: "Faroe.deletePasswordResetRequest()"
---

# Faroe.deletePasswordResetRequest()

Mapped to [DELETE /password-reset/\[request_id\]](/api-reference/rest/endpoints/delete_password-reset_requestid).

Deletes a passwod reset request. Deleting a non-existent request will not result in an error.

## Definition

```ts
async function deletePasswordResetRequest(requestId: string): Promise<void>
```

### Parameters

- `requestId`

## Error codes

- `UNKNOWN_ERROR`
