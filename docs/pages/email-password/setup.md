---
title: "Setting up your project"
---

# Setting up your project

For JavaScript projects, install the SDK and initialize it with server url, and if defined, the secret

```
npm install @faroe/sdk
```

```ts
// Without a secret.
export const faroe = new Faroe("http://localhost:4000", null);

// With a secret.
export const faroe = new Faroe("http://localhost:4000", secret);
```

In your application, you'll need to create a database table for your users, sessions, and password reset sessions.

In the user table, create a field for the Faroe user ID, user email, and an `email_verified` flag. You do not have to add a unique constraint to the email field.

```sql
CREATE TABLE user (
    id INTEGER NOT NULL PRIMARY KEY,
    faroe_id TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    email_verified INTEGER NOT NULL DEFAULT 0
);
```

Next, you'll need to implement sessions for managing the state of authenticated users. How you implement them is up to you but create an optional field for the Faroe email update request ID. For JavaScript projects, consider following the tutorial from [Lucia](https://lucia-auth.com).

```sql
CREATE TABLE session (
    id TEXT NOT NULL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user(id),
    expires_at INTEGER NOT NULL,
    faroe_email_update_request_id TEXT
);
```

Same goes for password reset session. This will be used maintain state during the reset flow. Password reset sessions should have an *absolute* expiration of 10 minutes, and have a field for the Faroe user ID, Faroe password reset request ID, and an `email_verified` flag. Optionally store your application's user ID.

```sql
CREATE TABLE password_reset_session (
    id TEXT NOT NULL PRIMARY KEY,
    faroe_user_id TEXT NOT NULL,
    faroe_request_id TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    email_verified INTEGER NOT NULL DEFAULT 0
);
```
