---
title: "Faroe.getUserPasswordResetRequests()"
---

# Faroe.getUserPasswordResetRequests()

Mapped to [GET /users/\[user_id\]/password-reset-requests](/reference/rest/endpoints/get_users_userid_password-reset-requests).

Gets an array of a user's valid password reset requests. Returns an empty array if the user doesn't have any valid reset requests or null if the user doesn't exist.

## Definition

```ts
//$ FaroePasswordResetRequest=/reference/sdk-js/main/FaroePasswordResetRequest
async function getUserPasswordResetRequests(
    userId: string
): Promise<$$FaroePasswordResetRequest[] | null>
```
## Error codes

- `UNKNOWN_ERROR`
