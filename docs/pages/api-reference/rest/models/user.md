---
title: "User model"
---

# User model

```ts
{
    "id": string,
    "created_at": number,
    "email": string,
    "recovery_code": string,
    "registered_totp": boolean
}
```

- `id`: A 24 character long unique identifier with 120 bits of entropy.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the user was created.
- `email`: A unique email that is 255 or less characters.
- `recovery_code`: A single-use code for resetting the user's second factors.
- `registered_totp`: `true` if the user holds a TOTP credential.

## Example

```json
{
    "id": "eeidmqmvdtjhaddujv8twjug",
    "created_at": 1728783738,
    "email": "user@example.com",
    "recovery_code": "12345678",
    "registered_totp": false
}
```
