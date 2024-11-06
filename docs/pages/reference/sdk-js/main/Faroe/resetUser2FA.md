---
title: "Faroe.resetUser2FA()"
---

# Faroe.resetUser2FA()

Mapped to [POST /users/\[user_id\]/reset-2fa](/reference/rest/endpoints/post_users_userid_reset-2fa).

Resets a user's second factors using a recovery code and returns a new recovery code. The user will be locked out from using their recovery code for 15 minutes after their 5th consecutive failed attempts.


## Definition

```ts
async function resetUser2FA(
    userId: string,
    recoveryCode: string
): Promise<string>;
```

## Parameters

- `userId`
- `recoveryCode`

## Error codes

- `TOO_MANY_REQUESTS`: Rate limit exceeded.
- `INCORRECT_CODE`: Incorrect recovery code.
- `NOT_FOUND`: The user does not exist.
- `INTERNAL_ERROR`
