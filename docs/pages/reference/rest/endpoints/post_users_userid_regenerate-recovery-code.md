---
title: "POST /users/[user_id]/regenerate-recovery-code"
---

# POST /users/[user_id]/regenerate-recovery-code

Regenerates a user's recovery code.

```
POST https://your-domain.com/users/USER_ID/regenerate-recovery-code
```

## Successful response

Return the user's new recovery code if the user exists.

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
- [500] `INTERNAL_ERROR`
