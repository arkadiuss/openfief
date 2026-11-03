-- Passkey checks and sign-out everywhere (docs/DESIGN.md D-183). Enrolling or removing a passkey, issuing an
-- enrollment link and creating a script token need a passkey check made in the same session within the last five
-- minutes, so each session records its last one, and the passkey that signed it in, whose removal revokes it.

-- A passkey check is a ceremony of its own kind, so a login's ceremony can't finish one, nor the other way round.
ALTER TABLE identity.ceremonies DROP CONSTRAINT ceremonies_kind_check;
ALTER TABLE identity.ceremonies ADD CONSTRAINT ceremonies_kind_check
    CHECK (kind IN ('registration', 'login', 'passkey-check'));

-- passkey_id: the credential ID of the passkey that signed the session in; no foreign key, since the session row
-- stays after its passkey is removed (and is revoked with it). passkey_checked_at: the session's last passkey check,
-- its sign-in or a later check. Sessions from before this migration have neither, so they need a check first.
ALTER TABLE identity.sessions
    ADD COLUMN passkey_id bytea CHECK (octet_length(passkey_id) BETWEEN 1 AND 1023),
    ADD COLUMN passkey_checked_at timestamptz CHECK (passkey_checked_at >= created_at),
    ADD CONSTRAINT sessions_passkey_only CHECK (authenticator = 'passkey' OR (passkey_id IS NULL AND passkey_checked_at IS NULL));

-- Removing a passkey revokes the sessions it signed in.
CREATE INDEX sessions_passkey_id ON identity.sessions (passkey_id) WHERE passkey_id IS NOT NULL;
