---
title: "TOTP credential model"
---

# TOTP credential model

```ts
{
    "id": string,
    "user_id": string,
    "created_at": number
}
```

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `user_id`: A 24-character long user ID.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the credential was created.

## Example

```json
{
    "id": "n5d6lm4scmfjv3jgurvj7xgq",
    "user_id": "vg6avv9dp7jvh36f8grjtpsj",
    "created_at": 1728783738
}
```
