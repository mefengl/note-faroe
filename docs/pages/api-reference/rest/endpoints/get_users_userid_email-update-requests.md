---
title: "GET /users/[user_id]/email-update-requests"
---

# GET /users/[user_id]/email-update-requests

Gets a list of a user's email update requests.

```
GET https://your-domain.com/users/USER_ID/email-update-requests
```

## Successful response

Returns a JSON array of [email update request models](/api-reference/rest/models/email-update-request). If the user does not have any update requests, it will return an empty array.

### Example

```json
[
    {
        "id": "dvd742g6mpmaebbjxq72kwsr",
        "user_id": "7six6i2igxd5ct4dccjk4qtg",
        "created_at": 1728803704,
        "expires_at": 1728804304,
        "email": "cat@example.com",
        "code": "VQ9REYBU"
    }
]
```

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
