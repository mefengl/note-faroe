---
title: "FaroeEmailVerificationRequest"
---

# FaroeEmailVerificationRequest

Mapped to the [email verification request model](/api-reference/rest/models/email-verification-request).

Represents an email verification request.

## Definition

```ts
interface FaroeEmailVerificationRequest {
	id: string;
	userId: string;
	createdAt: Date;
	expiresAt: Date;
	email: string;
	code: string;
}
```

### Properties

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `userId`: A 24-character long user ID.
- `created`: A timestamp representing when the request was created.
- `expiresAt`: A timestamp representing when the request will expire.
- `email`: An email that is 255 or less characters.
- `code`: An 8-character alphanumeric one-time code.
