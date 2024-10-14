---
title: "Faroe.deleteUserTOTPCredential()"
---

# Faroe.deleteUserTOTPCredential()

Mapped to [DELETE /users/\[user_id\]/totp](/api-reference/rest/endpoints/delete_users_userid_totp).

Deletes a user's TOTP credential. Deleting a non-existent credential will not result in an error.

## Definition

```ts
async function deleteUserTOTPCredential(
    userId: string,
    clientIP: string
): Promise<void>
```

### Parameters

- `userId`
- `clientIP`

## Error codes

- `UNKNOWN_ERROR`
