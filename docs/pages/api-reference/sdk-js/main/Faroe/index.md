---
title: "Faroe"
---

# Faroe

Represents a Faroe server client.

Server errors are thrown as [`FaroeError`](/api-reference/sdk-js/main/FaroeError). The error code is available from `FaroeError.code`. See each method for a list of possible error codes.

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

Errors caused by `fetch()` are wrapped as [`FaroeFetchError`](/api-reference/sdk-js/main/FaroeFetchError).

## Constructor

```ts
function constructor(url: string, credential: string | null): this
```

### Parameters

- `url`: The base URL of the Faroe server (e.g. `https://your-domain.com`).
- `credential`: The server credential.

## Methods

- [`createUser()`](/api-reference/sdk-js/main/Faroe/createUser)
- [`createUserEmailUpdateRequest()`](/api-reference/sdk-js/main/Faroe/createUserEmailUpdateRequest)
- [`createUserEmailVerificationRequest()`](/api-reference/sdk-js/main/Faroe/createUserEmailVerificationRequest)
- [`createUserPasswordResetRequest()`](/api-reference/sdk-js/main/Faroe/createUserPasswordResetRequest)
- [`deleteEmailUpdateRequest()`](/api-reference/sdk-js/main/Faroe/deleteEmailUpdateRequest)
- [`deletePasswordResetRequest()`](/api-reference/sdk-js/main/Faroe/deletePasswordResetRequest)
- [`deleteUser()`](/api-reference/sdk-js/main/Faroe/deleteUser)
- [`deleteUserEmailUpdateRequests()`](/api-reference/sdk-js/main/Faroe/deleteUserEmailUpdateRequests)
- [`deleteUserEmailVerificationRequest()`](/api-reference/sdk-js/main/Faroe/deleteUserEmailVerificationRequest)
- [`deleteUserPasswordResetRequests()`](/api-reference/sdk-js/main/Faroe/deleteUserPasswordResetRequests)
- [`deleteUserTOTPCredential()`](/api-reference/sdk-js/main/Faroe/deleteUserTOTPCredential)
- [`getEmailUpdateRequest()`](/api-reference/sdk-js/main/Faroe/getEmailUpdateRequest)
- [`getPasswordResetRequest()`](/api-reference/sdk-js/main/Faroe/getPasswordResetRequest)
- [`getUser()`](/api-reference/sdk-js/main/Faroe/getUser)
- [`getUserEmailUpdateRequests()`](/api-reference/sdk-js/main/Faroe/getUserEmailUpdateRequests)
- [`getUserEmailVerificationRequest()`](/api-reference/sdk-js/main/Faroe/getUserEmailVerificationRequest)
- [`getUserPasswordResetRequests()`](/api-reference/sdk-js/main/Faroe/getUserPasswordResetRequests)
- [`getUsers()`](/api-reference/sdk-js/main/Faroe/getUsers)
- [`getUserTOTPCredential()`](/api-reference/sdk-js/main/Faroe/getUserTOTPCredential)
- [`regenerateUserRecoveryCode()`](/api-reference/sdk-js/main/Faroe/regenerateUserRecoveryCode)
- [`registerUserTOTPCredential()`](/api-reference/sdk-js/main/Faroe/registerUserTOTPCredential)
- [`resetUser2FA()`](/api-reference/sdk-js/main/Faroe/resetUser2FA)
- [`resetUserPassword()`](/api-reference/sdk-js/main/Faroe/resetUserPassword)
- [`updateUserPassword()`](/api-reference/sdk-js/main/Faroe/updateUserPassword)
- [`verifyNewUserEmail()`](/api-reference/sdk-js/main/Faroe/verifyNewUserEmail)
- [`verifyPasswordResetRequestEmail()`](/api-reference/sdk-js/main/Faroe/verifyPasswordResetRequestEmail)
- [`verifyUser2FAWithTOTP()`](/api-reference/sdk-js/main/Faroe/verifyUser2FAWithTOTP)
- [`verifyUserEmail()`](/api-reference/sdk-js/main/Faroe/verifyUserEmail)
- [`verifyUserPassword()`](/api-reference/sdk-js/main/Faroe/verifyUserPassword)

## Example

```ts
import { Faroe } from "@faroe/sdk"

const faroe = new Faroe("https://your-domain.com", process.env.FAROE_CREDENTIAL);
```
