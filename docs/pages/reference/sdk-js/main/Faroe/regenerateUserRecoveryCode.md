---
title: "Faroe.regenerateUserRecoveryCode()"
---

# Faroe.regenerateUserRecoveryCode()

Mapped to [POST /users/\[user_id\]/regenerate-recovery-code](/reference/rest/endpoints/post_users_userid_regenerate-recovery-code).

Regenerates a user's recovery code and returns the new recovery code.

## Definition

```ts
async function regenerateUserRecoveryCode(
    userId: string
): Promise<string>
```

### Parameters

- `userId`

## Error codes

- `NOT_FOUND`
