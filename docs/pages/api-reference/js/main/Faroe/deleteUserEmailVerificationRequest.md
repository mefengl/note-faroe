---
title: "Faroe.deleteUserEmailVerificationRequest()"
---

# Faroe.deleteUserEmailVerificationRequest()

Mapped to [DELETE /users/\[user_id\]/email-verification/\[request_id\]](/api-reference/rest/endpoints/delete_users_userid_email-verification_requestid).

Deletes a user's email verification request. Deleting a non-existent request will not result in an error.

## Definition

```ts
async function deleteUserEmailVerificationRequest(
    userId: string,
    requestId: string
): Promise<void>
```

### Parameters

- `userId`
- `requestId`

## Error codes

- `UNKNOWN_ERROR`
