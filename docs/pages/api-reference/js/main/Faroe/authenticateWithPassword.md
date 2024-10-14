---
title: "Faroe.authenticateWithPassword()"
---

# Faroe.authenticateWithPassword()

Mapped to [POST /authenticate/password](/api-reference/rest/endpoints/post_authenticate_password).

Authenticates a user with email and password. It will temporary block the IP address if the client sends an incorrect password 5 times in a 15 minute window.

## Definition

```ts
//$ FaroeUser=/api-reference/js/main/FaroeUser
async function authenticateWithPassword(
    email: string,
    password: string,
    clientIP: string | null
): Promise<$$FaroeUser>
```

### Parameters

- `email`
- `password`
- `clientIP`

## Error codes

- `USER_NOT_EXISTS`: The user does not exist.
- `INCORRECT_PASSWORD`
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `UNKNOWN_ERROR`
