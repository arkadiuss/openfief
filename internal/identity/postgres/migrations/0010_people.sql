-- People (D-65, D-209, D-210): the owner and the viewers the owner invites. Each passkey and each session belongs
-- to a person; every one that exists before this migration is the owner's. A person is never deleted: a revoked
-- viewer keeps their row, with their ID and the history that names them.

CREATE TABLE identity.persons (
    -- The owner's ID is the word 'owner'; a viewer's is random, in hex, like a session's.
    id         text PRIMARY KEY CHECK (id = 'owner' OR id ~ '^[0-9a-f]{16}$'),
    kind       text NOT NULL CHECK (kind IN ('owner', 'viewer')),
    name       text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 64),
    -- A viewer's WebAuthn user handle; the owner's is identity.owner's.
    handle     bytea CHECK (octet_length(handle) = 32),
    created_at timestamptz NOT NULL,
    revoked_at timestamptz CHECK (revoked_at >= created_at),
    CHECK ((kind = 'owner') = (id = 'owner')),
    CHECK ((kind = 'viewer') = (handle IS NOT NULL)),
    CHECK (kind = 'viewer' OR revoked_at IS NULL)
);

-- Whoever signed in before is the owner, whether or not their row in identity.owner exists yet.
INSERT INTO identity.persons (id, kind, name, created_at)
VALUES ('owner', 'owner', 'Owner', COALESCE((SELECT created_at FROM identity.owner WHERE id = 1), now()));

ALTER TABLE identity.passkeys
    ADD COLUMN person_id   text NOT NULL DEFAULT 'owner' REFERENCES identity.persons (id),
    ADD COLUMN disabled_at timestamptz CHECK (disabled_at >= created_at);

CREATE INDEX passkeys_person ON identity.passkeys (person_id);

ALTER TABLE identity.sessions
    ADD COLUMN person_id text NOT NULL DEFAULT 'owner' REFERENCES identity.persons (id),
    -- A viewer's session only reads, so every call that changes something refuses it.
    ADD CONSTRAINT sessions_viewer_reads CHECK (person_id = 'owner' OR (scope = 'read' AND authenticator = 'passkey'));

CREATE INDEX sessions_person ON identity.sessions (person_id);

-- A link that invites a viewer carries what the owner chose for them: a name and the grants they start with.
ALTER TABLE identity.enrollment_codes
    ADD COLUMN viewer_name   text CHECK (char_length(viewer_name) BETWEEN 1 AND 64),
    ADD COLUMN viewer_grants jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(viewer_grants) = 'array'),
    ADD CONSTRAINT enrollment_codes_viewer_grants CHECK (viewer_name IS NOT NULL OR viewer_grants = '[]');

-- A registration begun with a viewer's link names the viewer it will create and the user handle the browser was given.
ALTER TABLE identity.ceremonies
    ADD COLUMN person_id     text CHECK (person_id ~ '^[0-9a-f]{16}$'),
    ADD COLUMN person_handle bytea CHECK (octet_length(person_handle) = 32),
    ADD CONSTRAINT ceremonies_person CHECK ((person_id IS NULL) = (person_handle IS NULL));

ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_actor_kind_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_actor_kind_check CHECK (actor_kind IN (
    'owner', 'viewer', 'enrollment_code', 'system', 'anonymous'));

ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call', 'model_call_result', 'person_created', 'person_revoked'));
