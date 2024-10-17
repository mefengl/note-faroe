---
title: "Faroe.deleteEmailUpdateRequest()"
---

# Faroe.deleteEmailUpdateRequest()

Mapped to [DELETE /email-update-requests/\[request_id\]](/api-reference/rest/endpoints/delete_email-update-requests_requestid).

Deletes an email update request. Deleting a non-existent request will not result in an error.

## Definition

```ts
async function deleteEmailUpdateRequest(requestId: string): Promise<void>
```

### Parameters

- `requestId`

## Error codes

- `UNKNOWN_ERROR`
