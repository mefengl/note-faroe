---
title: "Faroe.verifyUserEmail()"
---

# Faroe.verifyUserEmail()

Mapped to [POST /users/\[user_id\]/verify-email](/api-reference/rest/endpoints/post_users_userid_verify-email).

Verifies and updates a user's email with an email verification request code. Returns the verified email. Upon a successful verification, all email verification requests and password reset requests linked to the email are invalidated.

The verification request is immediately invalidated after the 5th failed attempt.

We recommend using the returned user data to update the user email of your app's database.

## Definition

```ts
async function verifyUserEmail(
    userId: string,
    requestId: string,
    code: string,
    clientIP: string | null
): Promise<string>
```

### Parameters

- `requestId`: An email verification request tied to the user.
- `code`: The one-time code of the email verification request.
- `clientIP`

## Error codes

- `INVALID_REQUEST_ID`: The request ID is invalid or has expired.
- `INCORRECT_CODE`: The one-time code is incorrect.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `UNKNOWN_ERROR`
