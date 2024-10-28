---
title: "FaroeUserTOTPCredential"
---

# FaroeUserTOTPCredential

Mapped to the [user totp credential model](/reference/rest/models/user-totp-credential).

Represents a user's totp credential.

## Definition

```ts
interface FaroeUserTOTPCredential {
	userId: string;
	createdAt: Date;
	key: Uint8Array;
}
```

### Properties

- `userId`: A 24-character long user ID.
- `createdAt`: A timestamp representing when the credential was created.
- `key`: A 20 byte key.
