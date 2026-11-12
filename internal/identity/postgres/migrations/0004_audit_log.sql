-- The audit log (docs/DESIGN.md D-211): an append-only record of every change of state and access event, each entry
-- naming who, what, which records by ID, when and the outcome, and never an amount, a secret or a question's text
-- (GR-12). Retention is forever. The columns are closed sets and IDs of a fixed shape, never free text, so the
-- database refuses what the domain (identity.AuditEntry) would refuse; adding an action is a migration.
CREATE TABLE identity.audit_log (
    id         bigserial PRIMARY KEY,
    at         timestamptz NOT NULL,
    actor_kind text NOT NULL CHECK (actor_kind IN ('owner', 'enrollment_code', 'system', 'anonymous')),
    actor_id   text NOT NULL CHECK (actor_id ~ '^([0-9a-f]{16}|core)?$'),
    action     text NOT NULL CHECK (action IN (
        'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
        'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked')),
    outcome    text NOT NULL CHECK (outcome IN ('ok', 'refused', 'failed')),
    -- A JSON array of {"kind": ..., "id": ...}: the records the entry names, by ID.
    refs       jsonb NOT NULL DEFAULT '[]' CHECK (
        jsonb_typeof(refs) = 'array' AND jsonb_array_length(refs) <= 8 AND octet_length(refs::text) <= 12000),
    CHECK ((actor_kind = 'anonymous') = (actor_id = '')),
    CHECK ((actor_kind = 'system') = (actor_id = 'core'))
);

-- The filters of the owner's view: by actor, by action and by time, each newest first.
CREATE INDEX audit_log_at ON identity.audit_log (at);
CREATE INDEX audit_log_actor ON identity.audit_log (actor_kind, actor_id, id DESC);
CREATE INDEX audit_log_action ON identity.audit_log (action, id DESC);

-- Append-only, enforced by the database in two layers. First, the role that runs this migration is the core's own
-- (on CloudNativePG, the cluster's app role, which owns what it creates), and it gives up UPDATE, DELETE and TRUNCATE
-- on the table, so the core can only INSERT and SELECT: a bug or an injected statement in the core fails with
-- "permission denied". Second, a trigger raises on any UPDATE, DELETE or TRUNCATE, so a role that still holds the
-- privilege, such as a superuser, is refused too. A table owner can still drop the trigger or the table, or grant
-- itself the privileges back: this is no defence against the database's own administrator, and no hash chain makes
-- tampering evident (a chain can start later from a genesis entry; P-18 is not decided).
REVOKE ALL ON identity.audit_log FROM PUBLIC;
REVOKE UPDATE, DELETE, TRUNCATE ON identity.audit_log FROM CURRENT_USER;

CREATE FUNCTION identity.audit_log_refuse() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'identity.audit_log is append-only: % refused', TG_OP USING ERRCODE = 'restrict_violation';
END
$$;

-- A statement trigger fires even when the statement matches no row.
CREATE TRIGGER audit_log_append_only
    BEFORE UPDATE OR DELETE OR TRUNCATE ON identity.audit_log
    FOR EACH STATEMENT EXECUTE FUNCTION identity.audit_log_refuse();
