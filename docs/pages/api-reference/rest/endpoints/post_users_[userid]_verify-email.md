---
title: "POST /users/[user_id]/verify-email"
---

# POST /users/[user_id]/verify-email

Verifies and updates a user's email with an email verification request code. Upon a successful verification, all email verification requests and password reset requests linked to the email are invalidated.

The verification request is immediately invalidated after the 5th failed attempt.

We recommend using the returned user data to update the user email of your app's database.

```
POST https://your-domain.com/users/USER_ID/verify-email
```

## Request body

All fields are required.

```ts
{
    "request_id": string,
    "code": string
}
```

- `request_id`: An email verification request tied to the user.
- `code`: The one-time code of the email verification request.

### Example

```json
{
    "request_id": "9sf9qcf3ctvwqwf9wdzw3fmj",
    "code": "9TW45AZU"
}
```

## Successful response

Returns the [user model](/api-reference/rest/models/user) of the email-verified user.

## Error codes

- [400] `INVALID_DATA`: Invalid request data.
- [400] `INVALID_REQUEST_ID`: The request ID is invalid or has expired.
- [400] `INCORRECT_CODE`: The one-time code is incorrect.
- [400] `TOO_MANY_REQUESTS`: Exceeded rate limit.
- [404] `NOT_FOUND`: The user does not exist.
- [500] `UNKNOWN_ERROR`
