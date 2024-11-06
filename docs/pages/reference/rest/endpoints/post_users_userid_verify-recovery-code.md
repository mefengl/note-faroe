---
title: "POST /users/[user_id]/verify-recovery-code"
---

# POST /users/[user_id]/verify-recovery-code

Verifies a user's recovery code and generates a new recovery code.

```
POST https://your-domain.com/users/USER_ID/verify-recovery-code
```

## Request body

All fields are required.

```json
{
    "recovery_code": string
}
```

- `recovery_code`

## Successful response

Return the user's new recovery code.

```json 
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

- [400] `INVALID_DATA`: Invalid request data.
- [400] `TOO_MANY_REQUESTS`: Rate limit exceeded.
- [400] `INCORRECT_CODE`: Incorrect recovery code.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `INTERNAL_ERROR`
