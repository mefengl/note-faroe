---
title: "Faroe.deleteUserEmailVerificationRequest()"
---

# Faroe.deleteUserEmailVerificationRequest()

Mapped to [DELETE /users/\[user_id\]/email-verification-request](/api-reference/rest/endpoints/delete_users_userid_email-verification-request).

Deletes a user's email verification request. Deleting a non-existent request will not result in an error.

## Definition

```ts
async function deleteUserEmailVerificationRequest(userId: string): Promise<void>
```

### Parameters

- `userId`

## Error codes

- `UNKNOWN_ERROR`
