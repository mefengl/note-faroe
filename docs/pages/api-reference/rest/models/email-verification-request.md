---
title: "Email verification request model"
---

# Email verificationn request model

```ts
{
    "id": string,
    "user_id": string,
    "created_at": number,
    "expires_at": number,
    "email": string,
    "code": string
}
```

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `user_id`: A 24-character long user ID.
- `created_at`: A 64-bit integer as an UNIX timestamp representing when the request was created.
- `expires_at`: A 64-bit integer as an UNIX timestamp representing when the request will expire.
- `email`: An email that is 255 or less characters.
- `code`: An 8-character alphanumeric one-time code.

## Example

```json
{
    "id": "9sf9qcf3ctvwqwf9wdzw3fmj",
    "user_id": "da7qg28mnk98nbyzwij5hsh7",
    "created_at": 1728803704,
    "expires_at": 1728804304,
    "email": "cat@example.com",
    "code": "9TW45AZU"
}
```
