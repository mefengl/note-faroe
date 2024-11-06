---
title: "Faroe.getUserEmailUpdateRequests()"
---

# Faroe.getUserEmailUpdateRequests()

Mapped to [GET /users/\[user_id\]/email-update-requests](/reference/rest/endpoints/get_users_userid_email-update-requests).

Gets an array of a user's valid email update requests. Returns an empty array if the user doesn't have any valid update requests or null if the user doesn't exist.

## Definition

```ts
//$ FaroeEmailUpdateRequest=/reference/sdk-js/main/FaroeEmailUpdateRequest
async function getUserEmailUpdateRequests(
    userId: string
): Promise<$$FaroeEmailUpdateRequest[] | null>
```
## Error codes

- `NOT_FOUND`: The user does not exist.
- `INTERNAL_ERROR`
