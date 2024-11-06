---
title: "User email verification request model"
---

# User email verification request model

```json
{
    "user_id": string,
    "created_at": number,
    "expires_at": number,
    "code": string
}
```

- `user_id`: A 24-character long user ID.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the request was created.
- `expires_at`: A 64-bit integer as an UNIX timestamp representing when the request will expire.
- `code`: An 8-character alphanumeric one-time code.

## Example

```json
{
    "user_id": "da7qg28mnk98nbyzwij5hsh7",
    "created_at": 1728803704,
    "expires_at": 1728804304,
    "code": "9TW45AZU"
}
```
