---
title: "POST /users/[user_id]/email-update-requests"
---

# POST /users/[user_id]/email-update-requests

Creates a new email update request for a user. This can only be called 3 times in a 15 minute window per user.

Send the created update request's code to the email address.

```
POST https://your-domain.com/users/USER_ID/email-update-requests
```

## Request body

```ts
{
    "email": string
}
```

- `email`: A valid email address.

## Successful response

Returns the [email update request model](/api-reference/rest/models/email-verification-request) of the created request.

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
