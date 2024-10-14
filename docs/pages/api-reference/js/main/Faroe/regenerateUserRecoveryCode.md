---
title: "Faroe.regenerateUserRecoveryCode()"
---

# Faroe.regenerateUserRecoveryCode()

Mapped to [POST /users/\[user_id\]/regenerate-recovery-code](/api-reference/rest/endpoints/post_users_userid_regenerate-recovery-code).

Regenerates a user's recovery code and returns the new recovery code.

## Definition

```ts
async function regenerateUserRecoveryCode(
    userId: string,
    clientIP: string | null
): Promise<string>
```

### Parameters

- `userId`
- `clientIP`

## Error codes

- `NOT_FOUND`
