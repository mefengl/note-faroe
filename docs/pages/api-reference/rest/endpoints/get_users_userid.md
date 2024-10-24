---
title: "GET /users/[user_id]"
---

# GET /users/[user_id]

Gets a user.

```
GET https://your-domain.com/users/USER_ID
```

## Successful response

Returns the [user model](/api-reference/rest/models/user) of the user if they exist. You can get the number of total pages from the `X-Pagination-Total` header.

```
X-Pagination-Total: 6
```

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
