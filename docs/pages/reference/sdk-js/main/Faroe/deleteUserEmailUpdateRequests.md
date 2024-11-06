---
title: "Faroe.deleteUserEmailUpdateRequests()"
---

# Faroe.deleteUserEmailUpdateRequests()

Mapped to [DELETE /users/\[user_id\]/email-update-requests](/reference/rest/endpoints/delete_users_userid_email-update-requests).

Deletes a user's email update requests. Attempting to delete email update requests of a non-existent user will not result in an error.

## Definition

```ts
async function deleteUserEmailUpdateRequests(
    userId: string
): Promise<void>
```
## Error codes

- `INTERNAL_ERROR`
