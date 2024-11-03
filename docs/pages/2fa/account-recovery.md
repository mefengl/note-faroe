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

Use `Faroe.verifyUserRecoveryCode()` to verify a user's recovery code. If valid, it will invalidate it and generate a new recovery code.

Set the session as not two-factor verified and delete the user's TOTP credential ID. Finally, use `Faroe.deleteUserSecondFactors()` to delete the Faroe user's second factors. The order is important here since we don't want a situation where updating your application's database fails and your database references deleted Faroe items.

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
    if (user.faroeUserTOTPCredentialId == null) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let code: string;

    // ...

    try {
        await faroe.verifyUserRecoveryCode(user.id, recoveryCode);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INCORRECT_CODE") {
            response.writeHeader(400);
            response.write("Incorrect code.");
            return;
        }
        if (e instanceof FaroeError && e.code === "TOO_MANY_REQUESTS") {
            response.writeHeader(429);
            response.write("Please try again later.");
            return;
        }
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
    }

    await setSessionAsNot2FAVerified(session.id);
    
    await deleteUserFaroeTOTPCredentialId(user.id);

    await faroe.deleteUserSecondFactors(user.faroeId);

    // ...
}
```

Use `Faroe.regenerateUserRecoveryCode()` to generate a new recovery code.

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
        recoveryCode = await faroe.regenerateUserRecoveryCode(user.id);
    } catch (e) {
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
    }

    // ...
}
```
