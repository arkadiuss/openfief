-- Passkeys and sessions (docs/DESIGN.md D-26, D-60; author decisions: auth split, demo read-only sessions).
-- Enrollment codes and sessions keep only the SHA-256 hash of their key, never the key (D-29).

-- The owner (GR-2): exactly one row, created with the first passkey or demo session. The handle is the WebAuthn
-- user ID every passkey carries.
CREATE TABLE identity.owner (
    id         smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    handle     bytea NOT NULL CHECK (octet_length(handle) = 32),
    created_at timestamptz NOT NULL
);

-- A passkey works only on its relying party ID, the host it was created on (D-60).
CREATE TABLE identity.passkeys (
    credential_id      bytea PRIMARY KEY CHECK (octet_length(credential_id) BETWEEN 1 AND 1023),
    rp_id              text NOT NULL CHECK (rp_id ~ '^[a-z0-9.-]{1,253}$'),
    name               text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 64),
    public_key         bytea NOT NULL CHECK (octet_length(public_key) > 0),
    attestation_type   text NOT NULL,
    attestation_format text NOT NULL,
    aaguid             bytea NOT NULL,
    sign_count         bigint NOT NULL CHECK (sign_count BETWEEN 0 AND 4294967295),
    transports         text[] NOT NULL,
    user_verified      boolean NOT NULL,
    backup_eligible    boolean NOT NULL,
    backup_state       boolean NOT NULL,
    created_at         timestamptz NOT NULL,
    last_used_at       timestamptz CHECK (last_used_at >= created_at)
);

-- One-time permissions to create a passkey on one origin.
CREATE TABLE identity.enrollment_codes (
    id         text PRIMARY KEY CHECK (id ~ '^[0-9a-f]{16}$'),
    origin     text NOT NULL,
    key_sha256 bytea NOT NULL CHECK (octet_length(key_sha256) = 32),
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
    used_at    timestamptz CHECK (used_at >= created_at)
);

-- WebAuthn ceremonies between begin and finish: deleted when finished, and expired ones when a new one begins.
CREATE TABLE identity.ceremonies (
    id                 text PRIMARY KEY CHECK (id ~ '^[0-9a-f]{32}$'),
    kind               text NOT NULL CHECK (kind IN ('registration', 'login')),
    origin             text NOT NULL,
    state              bytea NOT NULL,
    enrollment_code_id text REFERENCES identity.enrollment_codes (id),
    created_at         timestamptz NOT NULL,
    expires_at         timestamptz NOT NULL CHECK (expires_at > created_at)
);

-- Sessions: opaque, revocable, with an idle and an absolute expiry. Rows stay after logout or expiry.
CREATE TABLE identity.sessions (
    id            text PRIMARY KEY CHECK (id ~ '^[0-9a-f]{16}$'),
    scope         text NOT NULL CHECK (scope IN ('read', 'owner')),
    authenticator text NOT NULL CHECK (authenticator IN ('passkey', 'demo')),
    rp_id         text NOT NULL CHECK (rp_id ~ '^[a-z0-9.-]{1,253}$'),
    key_sha256    bytea NOT NULL CHECK (octet_length(key_sha256) = 32),
    created_at    timestamptz NOT NULL,
    last_seen_at  timestamptz NOT NULL CHECK (last_seen_at >= created_at),
    expires_at    timestamptz NOT NULL CHECK (expires_at > created_at),
    idle_seconds  bigint NOT NULL CHECK (idle_seconds > 0),
    revoked_at    timestamptz CHECK (revoked_at >= created_at),
    CHECK (authenticator = 'passkey' OR scope = 'read')
);
