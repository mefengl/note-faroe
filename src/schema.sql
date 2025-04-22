-- This file defines the database schema for the Faroe application using SQLite.
-- It creates tables to store user information, authentication details,
-- and various request types like email verification and password resets.

-- The 'user' table stores the core information for each registered user.
CREATE TABLE IF NOT EXISTS user (
    id TEXT NOT NULL PRIMARY KEY,           -- Unique identifier for the user (likely a generated string).
    created_at INTEGER NOT NULL,        -- Timestamp (Unix epoch seconds) when the user account was created.
    password_hash TEXT NOT NULL,        -- Securely hashed version of the user's password. NEVER store plain text passwords!
    recovery_code TEXT NOT NULL         -- A unique code provided to the user for account recovery (e.g., if they lose 2FA).
) STRICT; -- STRICT mode enforces data types more rigorously (e.g., INTEGER must be an integer).

-- The 'user_email_verification_request' table stores requests sent to users to verify their email address.
-- This is typically used right after registration.
CREATE TABLE IF NOT EXISTS user_email_verification_request (
    user_id TEXT NOT NULL UNIQUE PRIMARY KEY REFERENCES user(id), -- Links to the user who needs verification. UNIQUE ensures only one pending request per user.
    created_at INTEGER NOT NULL,        -- Timestamp when the verification request was created.
    expires_at INTEGER NOT NULL,        -- Timestamp when this verification request becomes invalid.
    code TEXT NOT NULL                  -- The secret code sent to the user's email for verification.
) STRICT;

-- The 'email_update_request' table stores requests made by users to change their registered email address.
-- This usually involves sending a verification code to the *new* email address.
CREATE TABLE IF NOT EXISTS email_update_request (
    id TEXT NOT NULL PRIMARY KEY,           -- Unique identifier for this specific update request.
    user_id TEXT NOT NULL REFERENCES user(id), -- Links to the user requesting the email change.
    created_at INTEGER NOT NULL,        -- Timestamp when the update request was created.
    expires_at INTEGER NOT NULL,        -- Timestamp when this update request becomes invalid.
    email TEXT NOT NULL,                -- The *new* email address the user wants to change to.
    code TEXT NOT NULL                  -- The secret code sent to the *new* email address for verification.
) STRICT;

-- Creates an index on the 'user_id' column of the 'email_update_request' table.
-- This speeds up looking up email update requests for a specific user.
CREATE INDEX IF NOT EXISTS email_update_request_user_id_index ON email_update_request(user_id);

-- The 'password_reset_request' table stores requests made by users to reset their password.
-- This typically involves sending a code or link to their verified email address.
CREATE TABLE IF NOT EXISTS password_reset_request (
    id TEXT NOT NULL PRIMARY KEY,           -- Unique identifier for this specific password reset request.
    user_id TEXT NOT NULL REFERENCES user(id), -- Links to the user requesting the password reset.
    created_at INTEGER NOT NULL,        -- Timestamp when the reset request was created.
    expires_at INTEGER NOT NULL,        -- Timestamp when this reset request becomes invalid.
    code_hash TEXT NOT NULL             -- A securely hashed version of the reset code sent to the user. Hashing prevents attackers from using stolen codes directly if the database is compromised.
) STRICT;

-- Creates an index on the 'user_id' column of the 'password_reset_request' table.
-- This speeds up looking up password reset requests for a specific user.
CREATE INDEX IF NOT EXISTS password_reset_request_user_id_index ON password_reset_request(user_id);

-- The 'user_totp_credential' table stores information related to Time-based One-Time Password (TOTP) setup for users (e.g., Google Authenticator).
CREATE TABLE IF NOT EXISTS user_totp_credential (
    user_id TEXT NOT NULL PRIMARY KEY REFERENCES user(id), -- Links to the user who has set up TOTP. PRIMARY KEY ensures only one TOTP setup per user.
    created_at INTEGER NOT NULL,        -- Timestamp when TOTP was set up for this user.
    key BLOB NULL                       -- The secret key shared between the server and the user's TOTP app. Stored as a binary large object (BLOB). NULL might indicate TOTP is not set up or temporarily disabled.
) STRICT;

-- The 'passkey_credential' table stores credentials for passwordless authentication using WebAuthn passkeys.
-- Passkeys allow users to log in using biometrics (fingerprint, face) or hardware keys, without a password.
CREATE TABLE IF NOT EXISTS passkey_credential (
    id TEXT NOT NULL,                   -- The unique credential ID provided by the browser/authenticator during registration. This is NOT the primary key for the *table* row itself.
    user_id TEXT NOT NULL REFERENCES user(id), -- Links to the user who owns this passkey.
    name TEXT NOT NULL,                 -- A user-friendly name for the passkey (e.g., "My Phone", "Work Laptop").
    created_at INTEGER NOT NULL,        -- Timestamp when the passkey was registered.
    cose_algorithm_id INTEGER NOT NULL, -- The COSE (CBOR Object Signing and Encryption) algorithm identifier used by this credential (e.g., ES256).
    public_key BLOB NULL                -- The public key part of the credential, stored as a binary large object. The corresponding private key is stored securely on the user's device.
) STRICT;

-- Creates an index on the 'user_id' column of the 'passkey_credential' table.
-- This speeds up looking up all passkeys registered by a specific user.
CREATE INDEX IF NOT EXISTS passkey_credential_user_id_index ON passkey_credential(user_id);

-- The 'security_key' table stores credentials for traditional FIDO/U2F security keys (a subset of WebAuthn).
-- Note: This table seems very similar to 'passkey_credential'. It might be for older U2F keys specifically,
-- or there might be a subtle difference in how they are handled compared to full passkeys.
-- The structure is identical to 'passkey_credential'.
CREATE TABLE IF NOT EXISTS security_key (
    id TEXT NOT NULL,                   -- The unique credential ID provided by the security key during registration.
    user_id TEXT NOT NULL REFERENCES user(id), -- Links to the user who owns this security key.
    name TEXT NOT NULL,                 -- A user-friendly name for the security key (e.g., "YubiKey").
    created_at INTEGER NOT NULL,        -- Timestamp when the security key was registered.
    cose_algorithm_id INTEGER NOT NULL, -- The COSE algorithm identifier used by this credential.
    public_key BLOB NULL                -- The public key part of the credential, stored as a binary large object.
) STRICT;

-- Creates an index on the 'user_id' column of the 'security_key' table.
-- This speeds up looking up all security keys registered by a specific user.
CREATE INDEX IF NOT EXISTS security_key_user_id_index ON security_key(user_id);
