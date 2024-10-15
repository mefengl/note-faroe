---
title: "Authenticator apps"
---

# Authenticator apps

*This page uses the JavaScript SDK*.

Implement 2FA with TOTP credentials to allow users to user their authenticator apps as their second factor.

## Register TOTP credential

We recommend install `@oslojs/otp` to create the key URI.

```
npm install @oslojs/otp
```

Generate a random 20 byte key and create a key URI with the interval set to 30 seconds and the digits to 6.

```ts
import { createTOTPKeyURI } from "@oslojs/otp";

const key = new Uint8Array(20);
crypto.getRandomValues(key);
const keyURI = createTOTPKeyURI("My app", user.email, key, 30, 6);
const qrcode = createQRCode(keyURI);
```

Ask the user to scan the QR code of the key with their authenticator app and enter the OTP code from the app. Send both the key (e.g. encode it with base64) and the code.

```ts
// HTTPRequest and HTTPResponse are just generic interfaces
async function handleRegisterTOTPRequest(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    const { session, user } = await validateRequest(request);
    if (session === null) {
        response.writeHeader(401);
        response.write("Not authenticated.");
        return;
    }

    if (user.registeredTOTP && !session.twoFactorVerified) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let key: Uint8Array;
    let code: string;

    // ...

    try {
        await faroe.registerUserTOTPCredential(user.faroeId, key, code, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INCORRECT_CODE") {
            response.writeHeader(400);
            response.write("Incorrect code.");
            return;
        }
        if (e instanceof FaroeError && e.code === "TOO_MANY_REQUESTS") {
            response.writeHeader(400);
            response.write("Please try again later.");
            return;
                }
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
    }

    await setUserAsRegisteredTOTP(user.id);
    await setSessionAs2FAVerified(session.id);

    // ...

}
```

## Verify TOTP

Use `Faroe.verifyUser2FAWithTOTP()` to verify TOTP codes.

If successful, set the session as `two_factor_verified`.

```ts
// HTTPRequest and HTTPResponse are just generic interfaces
async function handleVerifyUserTOTP(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    const { session, user } = await validateRequest(request);
    if (session === null) {
        response.writeHeader(401);
        response.write("Not authenticated.");
        return;
    }

    if (!user.registeredTOTP) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let code: string;

    try {
        await faroe.verifyUser2FAWithTOTP(user.faroeId, code, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INCORRECT_CODE") {
            response.writeHeader(400);
            response.write("Incorrect code.");
            return;
        }
        if (e instanceof FaroeError && e.code === "TOO_MANY_REQUESTS") {
            response.writeHeader(400);
            response.write("Please try again later.");
            return;
                }
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
    }

    await setSessionAs2FAVerified(session.id);

    // ...

}
```

2FA for password reset is nearly identical.

```ts
async function handleVerifyPasswordResetUserTOTP(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    const { session, user } = await validatePasswordResetRequest(request);
    if (session === null) {
        response.writeHeader(401);
        response.write("Not authenticated.");
        return;
    }

    if (!user.registeredTOTP) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let code: string;

    try {
        await faroe.verifyUser2FAWithTOTP(user.faroeId, code, clientIP);
    } catch (e) {
        // ...
    }

    await setPasswordResetSessionAs2FAVerified(session.id);

    // ...

}
```

## Disconnect TOTP credential

Use `Faroe.deleteUserTOTPCredential()` to delete a user's TOTP credential. Make sure that the user is 2FA-verified and to set `registered_totp` of the user to false afterwards.

```ts
// HTTPRequest and HTTPResponse are just generic interfaces
async function handleDisonnectTOTPCredential(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    const { session, user } = await validateRequest(request);
    if (session === null) {
        response.writeHeader(401);
        response.write("Not authenticated.");
        return;
    }

    if (!user.registeredTOTP) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }
    if (!session.twoFactorVerified) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    await faroe.deleteUserTOTPCredential(user.faroeId, clientIP);

    await setUserAsNotRegisteredTOTP(user.Id);

    // ...
}
```
