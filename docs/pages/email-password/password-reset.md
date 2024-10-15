---
title: "Password reset"
---

# Password reset

*This page uses the JavaScript SDK*.

## Send password reset email

Create a "forgot password" form and ask for the user's email. Create a new password reset request with `Faroe.createPasswordResetRequest()`. We recommend doing some basic input validation with `verifyEmailInput()`.

If successful, send the verification code to the user's inbox. Create a new password reset session and link the verification request to it.

We highly recommend putting some kind of bot and spam protection in front of this method.

```ts
import { verifyEmailInput, FaroeError } from "@faroe/sdk";

import type { FaroePasswordResetRequest } from "@faroe/sdk";

async function handleForgotPasswordRequest(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    let email: string;

    // ...

    if (!verifyEmailInput(email)) {
        response.writeHeader(400);
        response.write("Please enter a valid email address.");
        return;
    }

    let passwordResetRequest: FaroePasswordResetRequest;
    let verificationCode: string;
    try {
        [passwordResetRequest, verificationCode] = await faroe.createPasswordResetRequest(email, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "USER_NOT_EXISTS") {
            response.writeHeader(400);
            response.write("Account does not exist.");
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

    // Send verification code to user's inbox.
    const emailContent = `Your code is ${verificationCode}.`;
    await sendEmail(faroeUser.email, emailContent);

    const user = await getUserFromFaroeId(passwordResetRequest.userId);

    const passwordResetSession = await createPasswordResetSession(user.id, passwordResetRequest.Id, {
        emailVerified: false
    });

    // ...

}
```

## Verify password reset code

Ask the user for the verification code and use `Faroe.verifyPasswordResetRequestEmail()` to verify it. After the 5th failed attempt, the password reset request will be invalidated.

If successful, set the `email_verified` attribute of the session to `true`.

```ts
import { FaroeError } from "@faroe/sdk";

async function handleVerifyPasswordResetEmailRequest(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    const session = await validatePasswordResetRequest(request);
    if (session === null) {
        response.writeHeader(401);
        response.write("Not authenticated.");
        return;
    }

    let code: string;

    // ...

    try {
        await faroe.verifyPasswordResetRequestEmail(session.faroeRequestId, code, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INCORRECT_CODE") {
            response.writeHeader(400);
            response.write("Incorrect code.");
            return;
        }
        if (e instanceof FaroeError && e.code === "TOO_MANY_REQUESTS") {
            response.writeHeader(400);
            response.write("Please try again.");
            return;
        }
        if (e instanceof FaroeError && e.code === "NOT_FOUND") {
            await invalidatePasswordResetSession(session.id);
            response.writeHeader(400);
            response.write("Please restart the process.");
            return;
        }
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
    }

    await setPasswordResetSessionAsEmailVerified(session.id);

    // ...
}
```

## Reset password

Use `Faroe.resetPassword()` to reset the user's password using the password reset request. We recommend doing some basic input validation with `verifyPasswordInput()`.

**Ensure that the `email_verified` attribute of the password reset session is set to `true`.**

If successful, set the user's email as verified and invalidate all sessions belonging to the user.

```ts
import { verifyPasswordInput, FaroeError } from "@faroe/sdk";

async function handleResetPasswordRequest(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    const passwordResetSession = await validatePasswordResetRequest(request);
    if (passwordResetSession === null) {
        response.writeHeader(401);
        response.write("Not authenticated.");
        return;
    }
    // IMPORTANT!
    if (!passwordResetSession.emailVerified) {
        response.writeHeader(403);
        response.write("Forbidden.");
        return;
    }

    let password: string;

    // ...

    if (!verifyPasswordInput(password)) {
        response.writeHeader(400);
        response.write("Password must be 8 characters long.");
        return;
    }

    try {
        await faroe.resetPassword(session.faroeRequestId, password, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INVALID_REQUEST_ID") {
            response.writeHeader(400);
            response.write("Please restart the process.");
            return;
        }
        if (e instanceof FaroeError && e.code === "WEAK_PASSWORD") {
            response.writeHeader(400);
            response.write("Please use a stronger password.");
            return;
        }
        if (e instanceof FaroeError && e.code === "TOO_MANY_REQUESTS") {
            response.writeHeader(400);
            response.write("Please try again later.");
            return;
        }
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again.");
        return;
    }

    await setUserEmailAsVerified(passwordResetSession.userId);

    // Invalidate all sessions belonging to the user and create a new session.
    await invalidateUserSessions(passwordResetSession.userId);
    const session = await createSession(passwordResetSession.userId, null);

    // ...
}
```
