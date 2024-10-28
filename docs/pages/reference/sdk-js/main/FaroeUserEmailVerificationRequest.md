---
title: "FaroeUserEmailVerificationRequest"
---

# FaroeUserEmailVerificationRequest

Mapped to the [user email verification request model](/reference/rest/models/user-email-verification-request).

Represents an email verification request.

## Definition

```ts
interface FaroeUserEmailVerificationRequest {
	userId: string;
	createdAt: Date;
	expiresAt: Date;
	code: string;
}
```

### Properties

- `userId`: A 24-character long user ID.
- `created`: A timestamp representing when the request was created.
- `expiresAt`: A timestamp representing when the request will expire.
- `code`: An 8-character alphanumeric one-time code.
