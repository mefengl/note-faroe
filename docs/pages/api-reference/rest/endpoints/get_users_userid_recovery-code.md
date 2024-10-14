---
title: "GET /users/[user_id]/recovery-code"
---

# GET /users/[user_id]/recovery-code

Gets a user's recovery code.

```
GET https://your-domain.com/users/[user_id]/recovery-code
```

## Successful response

Return the user's recovery code if the user exists.

```ts
{
    "recovery_code": string
}
```

### Example

```json
{
    "recovery_code": "4UHZRTWP"
}
```

## Error codes

- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
