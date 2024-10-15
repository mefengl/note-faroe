---
title: "Update password"
---

# Update password

Use `Faroe.updateUserPassword()` to update the user's password using their current password. We recommend doing some basic input validation with `verifyPasswordInput()`. If successful, invalidate all existing sessions belonging to the user.

```ts
import { verifyPasswordInput, FaroeError } from "@faroe/sdk";

import type { FaroeUser } from "@faroe/sdk";

async function handleUpdatePasswordRequest(
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

    let password: string;
    let newPassword: string;
    // ...

    if (!verifyPasswordInput(newPassword)) {
        response.writeHeader(400);
        response.write("Password must be 8 characters long.");
        return;
    }

    try {
        await faroe.updateUserPassword(
            user.faroeId,
            password,
            newPassword.
            clientIP
        );
    } catch (e) {
        if (e instanceof FaroeError && e.code === "WEAK_PASSWORD") {
            response.writeHeader(400);
            response.write("Please use a stronger password.");
            return;
        }
        if (e instanceof FaroeError && e.code === "INCORRECT_PASSWORD") {
            response.writeHeader(400);
            response.write("Incorrect password.");
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

    // Invalidate all sessions belonging to the user and create a new session.
    await invalidateAllUserSessions(user.id);
    const session = createSession(user.id, null);

    // ...

}
```
