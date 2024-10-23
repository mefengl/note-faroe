---
title: "GET /users/[user_id]/password-reset-requests"
---

# GET /users/[user_id]/password-reset-requests

Gets a list of a user's valid password reset requests.

```
GET https://your-domain.com/users/USER_ID/password-reset-requests
```

## Successful response

Returns a JSON array of [password reset request models](/api-reference/rest/models/password-reset-request). If the user does not have any update requests, it will return an empty array.

### Example

```json
[
    {
        "id": "cjjhw9ggvv7e9hfc3qjsiegv",
        "user_id": "wz2nyjz4ims4cyuw7eq6tnxy",
        "created_at": 1728804201,
        "expires_at": 1728804801
    }
]
```

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
