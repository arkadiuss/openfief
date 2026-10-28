-- Script tokens (docs/DESIGN.md D-27). Only the SHA-256 hash of each token's key is stored, never the key or the
-- secret (D-29). Rows are never deleted: a revoked or expired token stays listed.
CREATE TABLE identity.script_tokens (
    id            text PRIMARY KEY CHECK (id ~ '^[0-9a-f]{16}$'),
    name          text NOT NULL CHECK (name ~ '^[A-Za-z0-9._-]{1,64}$'),
    scope         text NOT NULL CHECK (scope IN ('read')),
    key_sha256    bytea NOT NULL CHECK (octet_length(key_sha256) = 32),
    created_at    timestamptz NOT NULL,
    expires_at    timestamptz NOT NULL CHECK (expires_at > created_at),
    revoked_at    timestamptz CHECK (revoked_at >= created_at)
);
