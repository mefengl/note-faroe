---
title: "Faroe.createUserEmailUpdateRequest()"
---

# Faroe.createUserEmailUpdateRequest()

Mapped to [POST /users/\[user_id\]/email-update-requests](/reference/rest/endpoints/post_users_userid_email-update-requests).

Creates a new email verification request for a user. This can only be called 3 times in a 15 minute window per user.

## Definition

```ts
//$ $$FaroeEmailUpdateRequest=/reference/sdk-js/main/$$FaroeEmailUpdateRequest
async function createUserEmailUpdateRequest(
    userId: string,
    email: string
): Promise<$$FaroeEmailUpdateRequest>
```

### Parameters

- `userId`
- `email`: A valid email address.

## Error codes

- `INVALID_DATA`: Invalid email address.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `INTERNAL_ERROR`
