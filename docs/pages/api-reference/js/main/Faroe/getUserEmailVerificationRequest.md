---
title: "Faroe.getUserEmailVerificationRequest()"
---

# Faroe.getUserEmailVerificationRequest()

Mapped to [GET /users/\[user_id\]/email-verification-request](/api-reference/rest/endpoints/get_users_userid_email-verification-request).

Gets a user's email verification request. Returns `null` if the request doesn't exist or has expired.

## Definition

```ts
//$ FaroeUserEmailVerificationRequest=/api-reference/js/main/FaroeUserEmailVerificationRequest
async function getUserEmailVerificationRequest(
    userId: string
): Promise<$$FaroeUserEmailVerificationRequest | null>
```

### Parameters

- `userId`

## Error codes

- `UNKNOWN_ERROR`
