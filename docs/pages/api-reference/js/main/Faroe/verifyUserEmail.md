---
title: "Faroe.verifyUserEmail()"
---

# Faroe.verifyUserEmail()

Mapped to [POST /users/\[user_id\]/verify-email](/api-reference/rest/endpoints/post_users_userid_verify-email).

Verifies and updates a user's email with the user's email verification request code. The user will be locked out from verifying their email for 15 minutes after their 5th consecutive failed attempts.

## Definition

```ts
async function verifyUserEmail(
    userId: string,
    code: string,
    clientIP: string | null
): Promise<void>
```

### Parameters

- `userId`
- `code`: The one-time code of the email verification request.
- `clientIP`

## Error codes

- `NOT_ALLOWED`: The user does not have a request or has a expired request.
- `INCORRECT_CODE`: The one-time code is incorrect.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `UNKNOWN_ERROR`
