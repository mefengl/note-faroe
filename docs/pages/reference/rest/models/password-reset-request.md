---
title: "Password reset request model"
---

# Password reset request model

```json
{
    "id": string,
    "user_id": string,
    "created_at": number,
    "expires_at": number
}
```

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `user_id`: A 24-character long user ID.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the request was created.
- `expires_at`: A 64-bit integer as an UNIX timestamp representing when the request will expire.

## Example

```json
{
    "id": "cjjhw9ggvv7e9hfc3qjsiegv",
    "user_id": "wz2nyjz4ims4cyuw7eq6tnxy",
    "created_at": 1728804201,
    "expires_at": 1728804801
}
```
