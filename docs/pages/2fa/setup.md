---
title: "Setting up your project"
---

# Setting up your project

This section is based on the email and password guides.

## Update your database

Add a `two_factor_verified` attribute to your sessions.

```sql
CREATE TABLE session (
    id TEXT NOT NULL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user(id),
    expires_at INTEGER NOT NULL,
    faroe_email_verification_id TEXT,
    two_factor_verified INTEGER NOT NULL DEFAULT 0
);
```

## Update password reset flow

Add a `two_factor_verified` attribute to your password reset sessions.

```sql
CREATE TABLE password_reset_session (
    id TEXT NOT NULL PRIMARY KEY,
    faroe_user_id TEXT NOT NULL,
    faroe_request_id TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    email_verified INTEGER NOT NULL DEFAULT 0,
    two_factor_verified INTGEGER NOT NULL DEFAULT 0
);
```

If the user has a second factor registered, users should be prompted for it before they can reset their password.

```ts
const { session: passwordResetSession, user } = await validatePasswordResetRequest(request);
if (passwordResetSession === null) {
    response.writeHeader(401);
    response.write("Not authenticated.");
    return;
}

if (!passwordResetSession.emailVerified) {
    response.writeHeader(403);
    response.write("Forbidden.");
    return;
}
if (user.registeredTOTP && !session.twoFactorVerified) {
    response.writeHeader(403);
    response.write("Forbidden.");
    return;
}

// ...

try {
    await faroe.resetUserPassword(session.faroeRequestId, password, clientIP);
} catch (e) {
    // ...
}
```
