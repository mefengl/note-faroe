---
title: "Email verification"
---

# Email verification

*This page uses the JavaScript SDK*.

Ask the user for the email verification code sent to their inbox.

Get the email verification request linked to the current session and use `Faroe.verifyUserEmail()` to verify the user's email. If successful, set the `email_verified` field of your application's user to `true` and unlink the verification request from the session.

```ts
import { FaroeError } from "@faroe/sdk";

async function handleVerifyEmailRequest(
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

    if (user.emailVerified) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let code: string;

    // ...

    try {
        await faroe.verifyUserEmail(
            user.faroeId,
            code,
            clientIP
        );
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INVALID_REQUEST") {
            const emailVerificationRequest = await faroe.createUserEmailVerificationRequest(faroeUser.id, clientIP);
            const emailContent = `Your verification code is ${emailVerificationRequest.code}.`;
            await sendEmail(faroeUser.email, emailContent);

            response.writeHeader(400);
            response.write("Your verification code was expired. We sent a new one to your inbox.");
            return;
        }
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

    // Set email as verified.
    await setUserAsEmailVerified(session.userId);

    // Unlink verification request from session.
    await deleteSessionEmailVerificationRequestId(session.id);

    // ...

}
```

Like in the sign up process, use `Faroe.createUserEmailVerificationRequest()` to create a new email verification request. This method has rate limiting built-in to prevent DoS attacks targetting your email servers. However, consider adding some kind of bot and spam protection.

```ts
import { FaroeError } from "@faroe/sdk";

import type { FaroeEmailVerificationRequest } from "@faroe/sdk";

async function handleResendEmailVerificationCodeRequest(
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

    if (user.emailVerified) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let email: string;

    // ...

    if (!verifyEmailInput(email)) {
        response.writeHeader(400);
        response.write("Please enter a valid email address.");
        return;
    }

    let emailVerificationRequest: FaroeEmailVerificationRequest;
    try {
        emailVerificationRequest = await faroe.createUserEmailVerificationRequest(
            faroeUser.id,
            clientIP
        );
    } catch (e) {
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
    const emailContent = `Your verification code is ${emailVerificationRequest.code}.`;
    await sendEmail(faroeUser.email, emailContent);

    // Link the verification request to the current session.
    await setSessionEmailVerificationRequestId(session.id, emailVerificationRequest.id);

    // ...

}
```
