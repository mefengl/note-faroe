---
title: "Faroe.getUser()"
---

# Faroe.getPasswordResetRequest()

Mapped to [GET /users](/reference/rest/endpoints/get_users).

Gets a user. Returns `null` if the user doesn't exist.

## Definition

```ts
//$ FaroeUser=/reference/sdk-js/main/FaroeUser
async function getUser(userId: string): Promise<$$FaroeUser | null>
```

### Parameters

- `userId`

## Error codes

- `INTERNAL_ERROR`
