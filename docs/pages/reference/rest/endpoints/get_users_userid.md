---
title: "GET /users/[user_id]"
---

# GET /users/[user_id]

Gets a user.

```
GET https://your-domain.com/users/USER_ID
```

## Successful response

Returns the [user model](/reference/rest/models/user) of the user if they exist.

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
