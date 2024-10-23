---
title: "Faroe.deleteUserPasswordResetRequests()"
---

# Faroe.deleteUserPasswordResetRequests()

Mapped to [DELETE /users/\[user_id\]/password-reset-requests](/api-reference/rest/endpoints/delete_users_userid_email-reset-requests).

Deletes a user's password reset requests. Attempting to delete password reset requests of a non-existent user will not result in an error.

## Definition

```ts
async function deleteUserPasswordResetRequests(
    userId: string
): Promise<void>
```
## Error codes

- `UNKNOWN_ERROR`
