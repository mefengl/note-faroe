---
title: "FaroeUser"
---

# FaroeUser

Represents an user. Mapped to the [user model](/reference/rest/models/user).

## Definition

```ts
interface FaroeUser {
	id: string;
	createdAt: Date;
	recoveryCode: string;
	registeredTOTP: boolean;
}
```

### Properties

- `id`: A 24 character long unique identifier with 120 bits of entropy.
- `createdAt`: A timestamp representing when the user was created.
- `recoveryCode`: A single-use code for resetting the user's second factors.
- `registeredTOTP`: `true` if the user holds a TOTP credential.
