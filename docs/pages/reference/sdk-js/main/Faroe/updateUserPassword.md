---
title: "Faroe.updateUserPassword()"
---

# Faroe.updateUserPassword()

Mapped to [POST /users/\[user_id\]/password](/api-reference/rest/endpoints/post_users_userid_password).

Updates a user's password.

## Definition

```ts
async function updateUserPassword(
    userId: string,
    password: string,
    newPassword: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `userId`
- `password`: Current password.
- `new Password`: A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).
- `clientIP`

## Error codes

- `INVALID_DATA`: Invalid password length.
- `WEAK_PASSWORD`: The password is too weak.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `UNKNOWN_ERROR`
