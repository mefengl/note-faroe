CREATE TABLE IF NOT EXISTS user (
    id TEXT NOT NULL PRIMARY KEY,
    created_at INTEGER NOT NULL,
    email TEXT NOT NULL UNIQUE,
    email_verified INTEGER NOT NULL DEFAULT 0,
    password_hash TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS email_index ON user(email);

CREATE TABLE IF NOT EXISTS email_verification_request (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES user(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    email TEXT NOT NULL,
    code TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS email_verification_request_user_id_index ON email_verification_request(user_id);

CREATE TABLE IF NOT EXISTS password_reset_request (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES user(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    email TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    email_verified INTEGER NOT NULL DEFAULT 0,
    two_factor_verified INTEGER NOT NULL DEFAULT 0
) STRICT;

CREATE INDEX IF NOT EXISTS password_reset_request_user_id_index ON password_reset_request(user_id);

CREATE TABLE IF NOT EXISTS totp_credential (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE REFERENCES user(id),
    created_at INTEGER NOT NULL,
    key BLOB NULL
) STRICT;

CREATE INDEX IF NOT EXISTS totp_credential_user_id_index ON totp_credential(user_id);

CREATE TABLE IF NOT EXISTS passkey_credential (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES user(id),
    created_at INTEGER NOT NULL,
    cose_algorithm_id INTEGER NOT NULL,
    public_key BLOB NULL
) STRICT;

CREATE INDEX IF NOT EXISTS passkey_credential_user_id_index ON passkey_credential(user_id);

CREATE TABLE IF NOT EXISTS security_key (
    id TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES user(id),
    created_at INTEGER NOT NULL,
    cose_algorithm_id INTEGER NOT NULL,
    public_key BLOB NULL
) STRICT;

CREATE INDEX IF NOT EXISTS security_key_user_id_index ON security_key(user_id);
