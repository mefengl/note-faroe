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
	emailVerified: boolean;
	twoFactorVerified: boolean;
}
```

### Properties

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `userId`: A 24-character long user ID.
- `createdAt`: A timestamp representing when the request was created.
- `expiresAt`: A timestamp representing when the request will expire.
- `emailVerified`: `true` if the reset request's email was verified.
- `twoFactorVerified`: `true` if the second factor of the reset request's user was verified.
