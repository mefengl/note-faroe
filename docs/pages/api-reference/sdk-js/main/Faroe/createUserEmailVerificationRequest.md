---
title: "Faroe.createUserEmailVerificationRequest()"
---

# Faroe.createUserEmailVerificationRequest()

Mapped to [POST /users/\[user_id\]/email-verification-request](/api-reference/rest/endpoints/post_users_userid_email-verification-request).

Creates a new email verification request for a user. This can only be called 3 times in a 15 minute window per user.

## Definition

```ts
//$ FaroeUserEmailVerificationRequest=/api-reference/sdk-js/main/FaroeUserEmailVerificationRequest
async function createUserEmailVerificationRequest(
    userId: string
): Promise<$$FaroeUserEmailVerificationRequest>
```

### Parameters

- `userId`

## Error codes

- `INVALID_DATA`: Invalid email address.
- `EMAIL_ALREADY_USED`: The email is already used by an existing account.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `UNKNOWN_ERROR`
