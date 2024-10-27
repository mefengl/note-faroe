---
title: "Faroe.verifyUserPassword()"
---

# Faroe.verifyUserPassword()

Mapped to [POST /users/\[user_id\]/verify-password](/api-reference/rest/endpoints/post_users_userid_verify-password).

Verifies a user's password. It will temporary block the IP address if the client sends an incorrect password 5 times in a 15 minute window.

## Definition

```ts
async function verifyUserPassword(
    userId: string,
    password: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `userId`
- `password`
- `clientIP`

## Error codes

- `INCORRECT_PASSWORD`
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `UNKNOWN_ERROR`
- `NOT_FOUND`: User does not exist.
