---
title: "Sign up"
---

# Sign up

*This page uses the JavaScript SDK*.

Use `Faroe.createUser()` to register a new user. We recommend doing some basic input validation with `verifyEmailInput()` and `verifyPasswordInput()`. Pass the user's client IP address to enable IP-based rate limiting.

Next, create a new user for your application. Then, create a new email verification request with `Faroe.createUserEmailVerificationRequest()`. Send the verification code to the user's inbox and link the verification request to the current session.

We highly recommend putting some kind of bot and spam protection in front of this method.

```ts
// Everything not imported is something you need to define yourself.
import { verifyEmailInput, verifyPasswordInput, FaroeError } from "@faroe/sdk";

import type { FaroeUser } from "@faroe/sdk";

// HTTPRequest and HTTPResponse are just generic interfaces
async function handleSignUpRequest(
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
    
    // Check if email is already used.
    const existingUser = await getUserFromEmail(email);
    if (existingUser !== null) {
        response.writeHeader(400);
        response.write("Email is already used.");
        return;
    }

    if (!verifyPasswordInput(password)) {
        response.writeHeader(400);
        response.write("Password must be 8 characters long.");
        return;
    }
    

    let faroeUser: FaroeUser;
    try {
        faroeUser = await faroe.createUser(password, clientIP);
    } catch (e) {
        if (e instanceof FaroeError && e.code === "WEAK_PASSWORD") {
            response.writeHeader(400);
            response.write("Please use a stronger password.");
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

	let user: User;
	try {
		user = await createUser(faroeUser.id, email, {
            emailVerified: false
        });
	} catch {
		await faroe.deleteUser(faroeUser.id);
        response.writeHeader(500);
        response.write("An unknown error occurred. Please try again later.");
        return;
	}

    const emailVerificationRequest = await faroe.createUserEmailVerificationRequest(
        faroeUser.id,
        clientIP
    );
    const emailContent = `Your verification code is ${emailVerificationRequest.code}.`;
    await sendEmail(faroeUser.email, emailContent);

    // Create a session and link the verification request
    const session = await createSession(user.id, FaroeUserEmailVerificationRequest.id);

    // ...
}
```
