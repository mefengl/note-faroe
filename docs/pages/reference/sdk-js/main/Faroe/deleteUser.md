---
title: "Faroe.deleteUser()"
---

# Faroe.deleteUser()

Mapped to [DELETE /users/\[user_id\]](/reference/rest/endpoints/delete_users_userid).

Deletes a user. Deleting a non-existent user will not result in an error.

## Definition

```ts
async function deleteUser(userId: string): Promise<void>
```

### Parameters

- `userId`

## Error codes

- `INTERNAL_ERROR`
