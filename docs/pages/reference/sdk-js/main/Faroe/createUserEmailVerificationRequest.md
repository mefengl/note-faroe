---
title: "Faroe.createUserEmailVerificationRequest()"
---

# Faroe.createUserEmailVerificationRequest()

Mapped to [POST /users/\[user_id\]/email-verification-request](/reference/rest/endpoints/post_users_userid_email-verification-request).

Creates a new email verification request for a user. This can only be called 3 times in a 15 minute window per user.

## Definition

```ts
//$ FaroeUserEmailVerificationRequest=/reference/sdk-js/main/FaroeUserEmailVerificationRequest
async function createUserEmailVerificationRequest(
    userId: string
): Promise<$$FaroeUserEmailVerificationRequest>
```

### Parameters

- `userId`

## Error codes

- `INVALID_DATA`: Invalid email address.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `INTERNAL_ERROR`
