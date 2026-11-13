-- Grants (docs/DESIGN.md D-209): the owner's permission for a person, a script token or a view plugin to view one account,
-- one sleeve or the whole household. This table is the source of truth: SpiceDB holds a projection of the grants in force
-- (internal/identity/spicedb), written after the change commits and reconciled from this table at every start. A
-- grant only ever lets its holder view; nothing in a row says otherwise.
--
-- Revoking sets revoked_at and keeps the row, so the history of who could view what stays (D-210). A holder and a
-- target are named by ID alone, with the shapes identity.Grant validates, so the database refuses what the domain would
-- refuse. The account's or the sleeve's ID is a name in another context's schema (the ledger's, the portfolio's): there
-- is no foreign key, the core deletes a grant with the record it names, and the start sweeps one that outlived it.
CREATE TABLE identity.grants (
    id          bigserial PRIMARY KEY,
    holder_kind text NOT NULL CHECK (holder_kind IN ('person', 'script_token', 'view_plugin')),
    holder_id   text NOT NULL CHECK (
        (holder_kind IN ('person', 'script_token') AND holder_id ~ '^[0-9a-f]{16}$')
        OR (holder_kind = 'view_plugin' AND holder_id ~ '^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$')),
    target_kind text NOT NULL CHECK (target_kind IN ('account', 'sleeve', 'household')),
    target_id   text NOT NULL CHECK (
        (target_kind IN ('account', 'sleeve') AND target_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$')
        OR (target_kind = 'household' AND target_id = 'home')),
    created_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

-- At most one grant of a target to a holder is in force; any number have ended.
CREATE UNIQUE INDEX grants_in_force ON identity.grants (holder_kind, holder_id, target_kind, target_id)
    WHERE revoked_at IS NULL;
-- The grants of one holder, and of one target, for revoking with the record they name.
CREATE INDEX grants_holder ON identity.grants (holder_kind, holder_id);
CREATE INDEX grants_target ON identity.grants (target_kind, target_id);

-- The cutover to SpiceDB (D-209): when SpiceDB takes over, every script token and view plugin then holding read access
-- is given the household grant, once. This single row says it happened, so a token created later, with no grant yet,
-- is not given one by a later start.
CREATE TABLE identity.grant_cutover (
    only_row boolean PRIMARY KEY DEFAULT true CHECK (only_row),
    done_at  timestamptz NOT NULL
);

-- The audit log learns the two actions of a grant (D-211): grant_created and grant_revoked. The column's list is a
-- closed set, so adding an action is a migration, as 0004 says. This changes the table's definition, not a row, so the
-- append-only trigger and the withheld privileges are untouched.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked'));
