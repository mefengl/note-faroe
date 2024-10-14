---
title: "Faroe.getUserTOTPCredential()"
---

# Faroe.getUserTOTPCredential()

Mapped to [GET /users/\[user_id\]/totp](/api-reference/rest/endpoints/get_users_userid_totp).

Gets a user's TOTP credential. Returns `null` if the user doesn't exist or if the user doesn't have a TOTP credential.

## Definition

```ts
//$ FaroeUser=/api-reference/js/main/FaroeUser
async function getUserTOTPCredential(userId: string): Promise<$$FaroeUser | null>
```

### Parameters

- `userId`

## Error codes

- `UNKNOWN_ERROR`
