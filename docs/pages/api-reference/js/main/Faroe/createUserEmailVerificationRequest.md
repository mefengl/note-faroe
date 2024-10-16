---
title: "Faroe.createEmailVerificationRequest()"
---

# Faroe.createEmailVerificationRequest()

Mapped to [POST /users/\[user_id\]/email-verification-requests](/api-reference/rest/endpoints/post_users_userid_email-verification-requests).

Creates a new email verification request for a user. This can only be called 3 times in a 15 minute window per user.

Send the created verification request's code to the email address.

## Definition

```ts
//$ FaroeEamilVerificationRequest=/api-reference/js/main/FaroeEamilVerificationRequest
async function createEmailVerificationRequest(
    userId: string,
	email: string,
	clientIP: string | null
): Promise<$$FaroeEamilVerificationRequest>
```

### Parameters

- `email`: A valid email address.

## Error codes

- `INVALID_DATA`: Invalid email address.
- `EMAIL_ALREADY_USED`: The email is already used by an existing account.
- `TOO_MANY_REQUESTS`: Exceeded rate limit.
- `NOT_FOUND`: The user does not exist.
- `UNKNOWN_ERROR`
