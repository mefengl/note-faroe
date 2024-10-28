---
title: "Faroe"
---

# Faroe

Represents a Faroe server client.

Server errors are thrown as [`FaroeError`](/reference/sdk-js/main/FaroeError). The error code is available from `FaroeError.code`. See each method for a list of possible error codes.

```ts
import { Faroe, FaroeError } from "@faroe/sdk"

const faroe = new Faroe(url, secret);

try {
    await faroe.createUser(password, clientIP);
} catch (e) {
    if (e instanceof FaroeError) {
        const errorCode = e.code;
    }
}
```

Errors caused by `fetch()` are wrapped as [`FaroeFetchError`](/reference/sdk-js/main/FaroeFetchError).

## Constructor

```ts
function constructor(url: string, credential: string | null): this
```

### Parameters

- `url`: The base URL of the Faroe server (e.g. `https://your-domain.com`).
- `credential`: The server credential.

## Methods

- [`createUser()`](/reference/sdk-js/main/Faroe/createUser)
- [`createUserEmailUpdateRequest()`](/reference/sdk-js/main/Faroe/createUserEmailUpdateRequest)
- [`createUserEmailVerificationRequest()`](/reference/sdk-js/main/Faroe/createUserEmailVerificationRequest)
- [`createUserPasswordResetRequest()`](/reference/sdk-js/main/Faroe/createUserPasswordResetRequest)
- [`deleteEmailUpdateRequest()`](/reference/sdk-js/main/Faroe/deleteEmailUpdateRequest)
- [`deletePasswordResetRequest()`](/reference/sdk-js/main/Faroe/deletePasswordResetRequest)
- [`deleteUser()`](/reference/sdk-js/main/Faroe/deleteUser)
- [`deleteUserEmailUpdateRequests()`](/reference/sdk-js/main/Faroe/deleteUserEmailUpdateRequests)
- [`deleteUserEmailVerificationRequest()`](/reference/sdk-js/main/Faroe/deleteUserEmailVerificationRequest)
- [`deleteUserPasswordResetRequests()`](/reference/sdk-js/main/Faroe/deleteUserPasswordResetRequests)
- [`deleteUserTOTPCredential()`](/reference/sdk-js/main/Faroe/deleteUserTOTPCredential)
- [`getEmailUpdateRequest()`](/reference/sdk-js/main/Faroe/getEmailUpdateRequest)
- [`getPasswordResetRequest()`](/reference/sdk-js/main/Faroe/getPasswordResetRequest)
- [`getUser()`](/reference/sdk-js/main/Faroe/getUser)
- [`getUserEmailUpdateRequests()`](/reference/sdk-js/main/Faroe/getUserEmailUpdateRequests)
- [`getUserEmailVerificationRequest()`](/reference/sdk-js/main/Faroe/getUserEmailVerificationRequest)
- [`getUserPasswordResetRequests()`](/reference/sdk-js/main/Faroe/getUserPasswordResetRequests)
- [`getUsers()`](/reference/sdk-js/main/Faroe/getUsers)
- [`getUserTOTPCredential()`](/reference/sdk-js/main/Faroe/getUserTOTPCredential)
- [`regenerateUserRecoveryCode()`](/reference/sdk-js/main/Faroe/regenerateUserRecoveryCode)
- [`registerUserTOTPCredential()`](/reference/sdk-js/main/Faroe/registerUserTOTPCredential)
- [`resetUser2FA()`](/reference/sdk-js/main/Faroe/resetUser2FA)
- [`resetUserPassword()`](/reference/sdk-js/main/Faroe/resetUserPassword)
- [`updateUserPassword()`](/reference/sdk-js/main/Faroe/updateUserPassword)
- [`verifyNewUserEmail()`](/reference/sdk-js/main/Faroe/verifyNewUserEmail)
- [`verifyPasswordResetRequestEmail()`](/reference/sdk-js/main/Faroe/verifyPasswordResetRequestEmail)
- [`verifyUser2FAWithTOTP()`](/reference/sdk-js/main/Faroe/verifyUser2FAWithTOTP)
- [`verifyUserEmail()`](/reference/sdk-js/main/Faroe/verifyUserEmail)
- [`verifyUserPassword()`](/reference/sdk-js/main/Faroe/verifyUserPassword)

## Example

```ts
import { Faroe } from "@faroe/sdk"

const faroe = new Faroe("https://your-domain.com", process.env.FAROE_CREDENTIAL);
```
