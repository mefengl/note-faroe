---
title: "Update email address"
---

# Update email address

*This page uses the JavaScript SDK*.

The email verification requests used to verify the user's email can also be used to verify new email addresses.

```ts
import { verifyEmailInput, FaroeError } from "@faroe/sdk";

import type { FaroeEmailVerificationRequest } from "@faroe/sdk";

async function handleSendEmailVerificationCodeRequest(
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
            faroeUser.email,
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

After verifying the user, update your application's user's email address.

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

    if (session.faroeEmailVerificationRequestId === null) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let code: string;

    // ...

    let verifiedEmail: string
    try {
        verifiedEmail = await faroe.verifyUserEmail(
            user.faroeId,
            session.faroeEmailVerificationRequestId,
            code,
            clientIP
        );
    } catch (e) {

        // ...

    }

    await updateUserEmail(session.userId, verifiedEmail);

    // ...

}
```

For an improved user experience, you can use `Farore.getEmailVerificationRequest()` to check whether a request is still valid. Optionally, you can store the request expiration alongside your session.

```ts
if (session.faroeEmailVerificationRequestId === null) {
    response.writeHeader(403);
    response.write("Not allowed.");
    return;
}
const verificationRequest = await faroe.getEmailVerificationRequest(session.faroeEmailVerificationRequestId);
if (verificationRequest === null) {
    // Expired request.
    response.writeHeader(403);
    response.write("Not allowed.");
    return;
}
```
