---
title: "Account recovery"
---

# Account recovery

*This page uses the JavaScript SDK*.

Faroe supports recovery codes, which can be used to reset a user's second factors.

Use `Faroe.getUserRecoveryCode()` to get the user's recovery code. The code should displayed when the user first registers a second factor and should be accessible anytime after verifying their second factor.

```ts
const recoveryCode = await faroe.getUserRecoveryCode(faroeUserId);
```

Use `Faroe.resetUser2FA()` to reset the user's second factors with a recovery code. This will delete the user's TOTP credential and generate a new recovery code. Set the `registered_totp` user attribute to `false`.

```ts
import { FaroeError } from "@faroe/sdk";

import type { FaroeUser } from "@faroe/sdk";

async function handleReset2FARequest(
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

    // ...

    try {
        await faroe.resetUser2FA(user.id, recoveryCode, clientIP);
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

    await setUserAsNot2FARegistered(user.id);
    await setSessionAsNot2FAVerified(session.id);

    // ...
}
```

Use `Faroe.regenerateUserRecoveryCode()` to generate a new recovery code. Make sure that the user is 2FA-verified.

```ts
async function handleReset2FARequest(
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
    if (!session.twoFactorVerified) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let recoveryCode: string;
    try {
        recoveryCode = await faroe.regenerateUserRecoveryCode(user.id, clientIP);
    } catch (e) {
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
    }

    // ...
}
```
