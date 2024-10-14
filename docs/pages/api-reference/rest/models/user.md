---
title: "User model"
---

# User model

```ts
{
    "id": string,
    "created_at": number,
    "email": string,
    "email_verified": boolean,
    "registered_totp": boolean
}
```

- `id`: A 24 character long unique identifier with 120 bits of entropy.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the user was created.
- `email`: A unique email that is 255 or less characters.
- `email_verified`: `true` if the user's email was verified using a email verification request.
- `registered_totp`: `true` if the user holds a TOTP credential.

## Example

```json
{
    "id": "eeidmqmvdtjhaddujv8twjug",
    "created_at": 1728783738,
    "email": "user@example.com",
    "email_verified": true,
    "registered_totp": false
}
```
