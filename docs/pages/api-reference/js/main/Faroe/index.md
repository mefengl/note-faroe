---
title: "Faroe"
---

# Faroe

Represents a Faroe server client.

Server errors are thrown as [`FaroeError`](/api-reference/js/main/FaroeError). The error code is available from `FaroeError.code`. See each method for a list of possible error codes.

```ts
import { Faroe, FaroeError } from "@faroe/sdk"

const faroe = new Faroe(url, secret);

try {
    await faroe.createUser(username, password, clientIP);
} catch (e) {
    if (e instanceof FaroeError) {
        const errorCode = e.code;
    }
}
```

Errors caused by `fetch()` are wrapped as [`FaroeFetchError`](/api-reference/js/main/FaroeFetchError).

## Constructor

```ts
function constructor(url: string, credential: string): this
```

### Parameters

- `url`: The base URL of the Faroe server (e.g. `https://your-domain.com`).
- `credential`: The server credential.

## Methods

- [`authenticateWithPassword()`](/api-reference/js/main/Faroe/authenticateWithPassword)
- [`createPasswordResetRequest()`](/api-reference/js/main/Faroe/createPasswordResetRequest)
- [`createUser()`](/api-reference/js/main/Faroe/createUser)
- [`createUserEmailVerificationRequest()`](/api-reference/js/main/Faroe/createUserEmailVerificationRequest)
- [`deletePasswordResetRequest()`](/api-reference/js/main/Faroe/deletePasswordResetRequest)
- [`deleteUser()`](/api-reference/js/main/Faroe/deleteUser)
- [`deleteUserEmailVerificationRequest()`](/api-reference/js/main/Faroe/deleteUserEmailVerificationRequest)
- [`deleteUserTOTPCredential()`](/api-reference/js/main/Faroe/deleteUserTOTPCredential)
- [`getPasswordResetRequest()`](/api-reference/js/main/Faroe/getPasswordResetRequest)
- [`getUser()`](/api-reference/js/main/Faroe/getUser)
- [`getUserEmailVerificationRequest()`](/api-reference/js/main/Faroe/getUserEmailVerificationRequest)
- [`getUsers()`](/api-reference/js/main/Faroe/getUsers)
- [`getUserTOTPCredential()`](/api-reference/js/main/Faroe/getUserTOTPCredential)
- [`regenerateUserRecoveryCode()`](/api-reference/js/main/Faroe/regenerateUserRecoveryCode)
- [`registerUserTOTPCredential()`](/api-reference/js/main/Faroe/registerUserTOTPCredential)
- [`resetUser2FA()`](/api-reference/js/main/Faroe/resetUser2FA)
- [`resetUserPassword()`](/api-reference/js/main/Faroe/resetUserPassword)
- [`updateUserPassword()`](/api-reference/js/main/Faroe/updateUserPassword)
- [`verifyPasswordResetRequestEmail()`](/api-reference/js/main/Faroe/verifyPasswordResetRequestEmail)
- [`verifyUser2FAWithTOTP()`](/api-reference/js/main/Faroe/verifyUser2FAWithTOTP)
- [`verifyUserEmail()`](/api-reference/js/main/Faroe/verifyUserEmail)

## Example

```ts
import { Faroe } from "@faroe/sdk"

const faroe = new Faroe("https://your-domain.com", process.env.FAROE_CREDENTIAL);
```
