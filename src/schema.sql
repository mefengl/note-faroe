CREATE TABLE IF NOT EXISTS user (
    id TEXT NOT NULL PRIMARY KEY,
    created_at INTEGER NOT NULL,
    password_hash TEXT NOT NULL,
    recovery_code TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS user_email_verification_request (
    user_id TEXT NOT NULL UNIQUE PRIMARY KEY REFERENCES user(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    code TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS email_update_request (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES user(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    email TEXT NOT NULL,
    code TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS email_update_request_user_id_index ON email_update_request(user_id);

CREATE TABLE IF NOT EXISTS password_reset_request (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES user(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    code_hash TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS password_reset_request_user_id_index ON password_reset_request(user_id);

CREATE TABLE IF NOT EXISTS user_totp_credential (
    user_id TEXT NOT NULL PRIMARY KEY REFERENCES user(id),
    created_at INTEGER NOT NULL,
    key BLOB NULL
) STRICT;

CREATE TABLE IF NOT EXISTS passkey_credential (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES user(id),
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    cose_algorithm_id INTEGER NOT NULL,
    public_key BLOB NULL
) STRICT;

CREATE INDEX IF NOT EXISTS passkey_credential_user_id_index ON passkey_credential(user_id);

CREATE TABLE IF NOT EXISTS security_key (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES user(id),
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    cose_algorithm_id INTEGER NOT NULL,
    public_key BLOB NULL
) STRICT;

CREATE INDEX IF NOT EXISTS security_key_user_id_index ON security_key(user_id);
