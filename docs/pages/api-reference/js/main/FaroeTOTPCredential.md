---
title: "FaroeTOTPCredential"
---

# FaroeTOTPCredential

Mapped to the [password reset request model](/api-reference/rest/models/password-reset-request).

Represents an user.

## Definition

```ts
interface FaroeTOTPCredential {
	id: string;
	userId: string;
	createdAt: Date;
	key: Uint8Array;
}
```

### Properties

- `id`: A 24-character long unique identifier with 120 bits of entropy.
- `userId`: A 24-character long user ID.
- `createdAt`: A timestamp representing when the credential was created.
- `key`: A 20 byte key.
