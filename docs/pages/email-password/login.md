---
title: "Sign in"
---

# Sign in

*This page uses the JavaScript SDK*.

Get the user from your database with the email and use `Faroe.verifyUserPassword()` to authenticate a user with email and password. We recommend doing some basic input validation with `verifyEmailInput()` and `verifyPasswordInput()`. Pass the user's client IP address to enable IP-based rate limiting.

If successful, get the user from the Faroe user ID and create a new session.

```ts
// Everything not imported is something you need to define yourself.
import { verifyEmailInput, FaroeError } from "@faroe/sdk";

import type { FaroeUser } from "@faroe/sdk";

async function handleLoginRequest(
    request: HTTPRequest,
    response: HTTPResponse
): Promise<void> {
    const clientIP = request.headers.get("X-Forwarded-For");

    let email: string;
    let password: string;
    // ...

    if (!verifyEmailInput(email)) {
        response.writeHeader(400);
        response.write("Please enter a valid email address.");
        return;
    }

    const user = await getUserFromEmail(email);
    if (user === null) {
        response.writeHeader(400);
        response.write("Account does not exist.");
        return;
    }

    let faroeUser: FaroeUser;
    try {
        faroeUser = await faroe.verifyUserPassword(user.faroeId, password, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "INCORRECT_PASSWORD") {
            response.writeHeader(400);
            response.write("Incorrect password.");
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

    const session = await createSession(user.id, null);

    // ...
}
```
