---
title: "Faroe.createPasswordResetRequest()"
---

# Faroe.createPasswordResetRequest()

Mapped to [POST /password-reset](/api-reference/rest/endpoints/post_password-reset).

Creates a new password reset request for a user. This can only be called 3 times in a 15 minute window per user.

Send the created reset request's code to the email address.

## Definition

```ts
//$ FaroePasswordResetRequest=/api-reference/js/main/$$FaroePasswordResetRequest
async function createPasswordResetRequest(
    email: string,
    clientIP: string | null
): Promise<[request: $$FaroePasswordResetRequest, code: string]>
```

### Parameters

- `email`: A valid email address.
- `clientIP`

## Error codes

- `INVALID_DATA`: Malformed email address.
- `USER_NOT_EXISTS`: A user linked to the email does not exist.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `UNKNOWN_ERROR`
