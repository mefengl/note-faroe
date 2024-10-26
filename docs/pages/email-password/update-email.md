---
title: "Update email address"
---

# Update email address

*This page uses the JavaScript SDK*.

Create a new email update request, send the verification code to the user's inbox, and link the update request to the current session.

```ts
import { verifyEmailInput, FaroeError } from "@faroe/sdk";

import type { FaroeEmailUpdateRequest } from "@faroe/sdk";

async function handleSendEmailUpdateVerificationCodeRequest(
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

    let emailUpdateRequest: FaroeEmailUpdateRequest;
    try {
        emailUpdateRequest = await faroe.createUserEmailUpdateRequest(
            faroeUser.id,
            faroeUser.email,
            clientIP
        );
    } catch (e) {
        if (e instanceof FaroeError && e.code === "EMAIL_ALREADY_USED") {
            response.writeHeader(400);
            response.write("This email address is already used.");
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

    // Send verification code to user's inbox.
    const emailContent = `Your verification code is ${emailUpdateRequest.code}.`;
    await sendEmail(faroeUser.email, emailContent);

    // Link the verification request to the current session.
    await setSessionEmailUpdateRequestId(session.id, emailUpdateRequest.id);

    // ...

}
```

Verify the code and update your application's user's email address.

```ts
import { FaroeError } from "@faroe/sdk";

async function handleUpdateEmailRequest(
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

    if (session.faroeEmailUpdateRequestId === null) {
        response.writeHeader(403);
        response.write("Not allowed.");
        return;
    }

    let code: string;

    // ...

    let verifiedEmail: string
    try {
        verifiedEmail = await faroe.updateUserEmail(
            session.faroeEmailUpdateRequestId,
            code,
            clientIP
        );
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INVALID_REQUEST") {
            response.writeHeader(400);
            response.write("Please restart the process.");
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

    await updateUserEmail(session.userId, verifiedEmail);

    await deleteSessionEmailUpdateRequestId(session.id);

    // ...

}
```
