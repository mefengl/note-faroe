# Faroe

*This software is not stable yet. Do not use it in production.*

Faroe is an open source, self-hosted, and modular authentication backend designed for email and password authentication. It exposes various API endpoints including:

- Registering users with email and password
- Authenticating users with email and password
- Email verification
- Password reset
- 2FA with TOTP
- 2FA recovery

These work with your application's UI and backend to provide a complete authentication system.

```ts
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

Faroe is written in Go and uses SQLite as its database.

Licensed under the MIT license.

## Features

- Email login, email verification, 2FA with TOTP, 2FA recovery, and password reset
- Rate limiting and brute force protection
- Proper password strength checks
- Everything included in a single binary
- WIP: Logging
- WIP: Dashboard
- WIP: Database backup
- WIP: WebAuthn

## Why?

If you don't want to use a fullstack framework, implementing auth means paying for a third-party service, self-hosting an identity provider, or building one from scratch. JavaScript especially is yet to have a standard, default framework a built-in auth solution. A separate backend that handles everything is nice, but it can be frustrating to customize the overall login flow, data structure, and UI. Implementing from scratch gives you the most flexibility, but becomes time-consuming when you want to implement anything more than OAuth.

Faroe is the middle ground between a dedicated auth backend and a custom implementation. You can let it handle the core logic and just build the UI and manage sessions. It's most of the hard part of auth compressed into a single binary file.

## Things to consider

- Faroe does not include an email server.
- Bot protection is not included. We highly recommend using Captchas or equivalent in registration and password reset forms.
- Faroe uses SQLite in WAL mode as its database. This shouldn't cause issues unless you have 100,000+ users, and even then, the database will only handle a small part of your total requests.
- Faroe uses in-memory storage for rate limiting.
