---
title: "Faroe.registerUserTOTPCredential()"
---

# Faroe.registerUserTOTPCredential()

Mapped to [POST /users/\[user_id\/totp](/api-reference/rest/endpoints/post_users_userid_totp).

Registers a TOTP (SHA-1, 6 digits, 30 seconds interval) credential to a user.

## Definition

```ts
//$ FaroeUserTOTPCredential=/api-reference/js/main/FaroeUserTOTPCredential
async function registerUserTOTPCredential(
    userId: string,
    totpKey: Uint8Array,
    code: string
): Promise<$$FaroeUserTOTPCredential>
```

### Parameters

- `userId`
- `totp_key`: A base64-encoded TOTP key. The encoded key must be 20 bytes.
- `code`: The TOTP code from the key for verification.

## Error codes

- `INVALID_DATA`: Invalid TOTP key length.
- `INCORRECT_CODE`
- `NOT_FOUND`: Invalid user ID.
- `UNKNOWN_ERROR`

### Example

```ts
import { Faroe, UserSortBy, SortOrder } from "@faroe/sdk";

const faroe = new Faroe(url, secret);

const key = new Uint8Array(20);
crypto.getRandomValues(key);

// ...

const credential = await faroe.registerUserTOTPCredential(userId, key, code);
```
