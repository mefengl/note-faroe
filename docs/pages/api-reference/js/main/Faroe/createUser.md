---
title: "Faroe.createUser()"
---

# Faroe.createUser()

Mapped to [POST /users](/api-reference/rest/endpoints/post_users).

Creates a new user.

We highly recommend putting a Captcha or equivalent in front for spam and bot detection.

## Definition

```ts
//$ FaroeUser=/api-reference/js/main/FaroeUser
async function createUser(
    email: string,
    password: string,
    clientIP: string | null
): Promise<$$FaroeUser>
```

### Parameters

- `email`: A valid email address.
- `password`: A valid password. Password strength is determined by checking it aginst past data leaks using the [HaveIBeenPwned API](https://haveibeenpwned.com/API/v3#PwnedPasswords).
- `clientIP`

## Error codes

- `INVALID_DATA`: Malformed email address; invalid email address or password length.
- `EMAIL_ALREADY_USED`
- `WEAK_PASSWORD`: The password is too weak.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `UNKNOWN_ERROR`
