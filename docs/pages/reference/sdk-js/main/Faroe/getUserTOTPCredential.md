---
title: "Faroe.getUserTOTPCredential()"
---

# Faroe.getUserTOTPCredential()

Mapped to [GET /users/\[user_id\]/totp-credential](/reference/rest/endpoints/get_users_userid_totp-credential).

Gets a user's TOTP credential. Returns `null` if the user doesn't exist or if the user doesn't have a TOTP credential.

## Definition

```ts
//$ FaroeUserTOTPCredential=/reference/sdk-js/main/FaroeUserTOTPCredential
async function getUserTOTPCredential(userId: string): Promise<$$FaroeUserTOTPCredential | null>
```

### Parameters

- `userId`

## Error codes

- `INTERNAL_ERROR`
