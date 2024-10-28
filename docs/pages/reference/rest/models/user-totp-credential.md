---
title: "User TOTP credential model"
---

# User TOTP credential model

```ts
{
    "user_id": string,
    "created_at": number,
    "key": string
}
```

- `user_id`: A 24-character long user ID.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the credential was created.
- `key`: A base64-encoded 20 byte key.

## Example

```json
{
    "user_id": "vg6avv9dp7jvh36f8grjtpsj",
    "created_at": 1728783738,
    "key": "nHKsL9EFvdzuTWMzGCjZgZWojpU="
}
```
