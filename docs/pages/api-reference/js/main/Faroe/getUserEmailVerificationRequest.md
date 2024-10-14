---
title: "Faroe.getUserEmailVerificationRequest()"
---

# Faroe.getUserEmailVerificationRequest()

Mapped to [GET /users/\[user_id\]/email-verification/\[request_id\]](/api-reference/rest/endpoints/get_users_userid_email-verification_requestid).

Gets a email verification request. Returns `null` if the user or request doesn't exist.

## Definition

```ts
//$ FaroeEmailVerificationRequest=/api-reference/js/main/FaroeEmailVerificationRequest
async function getUserEmailVerificationRequest(
    userId: string,
    requestId: string
): Promise<$$FaroeEmailVerificationRequest | null>
```

### Parameters

- `userId`
- `requestId`

## Error codes

- `UNKNOWN_ERROR`
