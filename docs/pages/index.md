---
title: "Faroe"
---

# Faroe

Faroe is an open source, self-hosted, and modular identity provider specifically for email and password authentication. It exposes various API endpoints including:

- Registering users with email and password
- Authenticating users with email and password
- Email verification
- Password reset
- 2FA with TOTP
- 2FA recovery

These work with your application's UI and backend to provide a complete authentication system.

```ts
// Using the JavaScript SDK.
import { Faroe } from "@faroe/sdk"

const faroe = new Faroe(url, secret);

async function handleLoginRequest() {

  // ...

  const faroeUser = await faroe.withIP(clientIP).authenticateWithPassword(email, password);

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

- Email login, email verification, 2FA with TOTP, 2FA recovery, and password reset
- Rate limiting and brute force protection
- Proper password strength checks
- WIP: Logging
- WIP: Dashboard
- WIP: Database backup
- WIP: WebAuthn
