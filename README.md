# Faroe

*This software is not stable yet. Do not use it in production.*

Faroe is an open source, self-hosted, and modular identity provider specifically for email and password authentication. It exposes various API endpoints including:

- Registering users with email and password
- Authenticating users with email and password
- Email verification
- Password reset
- 2FA with TOTP

These work with your application's UI and backend to provide a complete authentication system.

```ts
import { Client } from "@faroe/sdk"

const client = new Client(url, secret);

async function handleLoginRequest() {

  // ...

  const faroeUser = await client.withIP(clientIP).authenticateWithPassword(email, password);

  // Your application logic
  const user = await getUserByFaroeId(faroeUser.id);
  const session = await createSession(user.id, {
    twoFactorVerified: false
  });
}
```

This is not a full authentication backend (Auth0, Supabase, etc) nor a full identity provider (KeyCloak, etc). It is specfically designed to only handle the backend logic for email and password authentication. Faroe does not provide session management, frontend UI, or OAuth integration.

Faroe is written in GO and uses SQLite as its database.

Licensed under the MIT license.

## Features

- Email login, email verification, 2FA with TOTP, and password reset
- Rate limiting and brute force protection
- Proper password strength checks
- WIP: Logging
- WIP: Dashboard
- WIP: Database backup
- WIP: WebAuthn

## Things to consider

- Faroe does not include an email server.
- Bot protection is not included. We highly recommend using Captchas or equivalent in registration and password reset forms.
- Faroe uses SQLite in WAL mode as its database. This shouldn't cause issues unless you have 100,000+ users, and even then, the database will only handle a small part of your total requests.
- Faroe uses in-memory storage for rate limiting.
