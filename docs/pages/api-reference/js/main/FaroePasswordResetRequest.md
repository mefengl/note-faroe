---
title: "FaroePasswordResetRequest"
---

# FaroePasswordResetRequest

Mapped to the [password reset request model](/api-reference/rest/models/password-reset-request).

Represents an user.

## Definition

```ts
interface FaroePasswordResetRequest {
	id: string;
	userId: string;
	createdAt: Date;
	expiresAt: Date;
	email: string;
	emailVerified: boolean;
	twoFactorVerified: boolean;
}
```

### Properties

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `userId`: A 24-character long user ID.
- `createdAt`: A timestamp representing when the request was created.
- `expiresAt`: A timestamp representing when the request will expire.
- `email`: An email that is 255 or less characters.
- `emailVerified`: `true` if the reset request's email was verified.
- `twoFactorVerified`: `true` if the second factor of the reset request's user was verified.
