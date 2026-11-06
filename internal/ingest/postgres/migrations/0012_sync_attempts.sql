-- The history of each connector instance's sync attempts (docs/DESIGN.md D-49, D-196): when each attempt started and
-- ended, which batch it ran, and for a failed one only its error class, never the error's text, which may quote an
-- amount, an ISIN or an account number (GR-12); with how many attempts in a row had failed and when the next was due.
-- A sync that keeps failing raises an attention item that shows this timeline. Rows are only appended.
CREATE TABLE ingest.sync_attempts (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    instance     text NOT NULL CHECK (instance ~ '^[a-z]([-a-z0-9]*[a-z0-9])?$' AND length(instance) <= 56),
    batch_id     text NOT NULL REFERENCES ingest.import_batches (id),
    mode         text NOT NULL CHECK (mode IN ('opening', 'transactions')),
    started_at   timestamptz NOT NULL,
    ended_at     timestamptz NOT NULL CHECK (ended_at >= started_at),
    -- NULL for a successful attempt.
    error_class  text CHECK (error_class IN (
        'timeout', 'unavailable', 'connector-error', 'invalid-delivery', 'refused', 'booking', 'interrupted', 'other')),
    failures     integer NOT NULL CHECK (failures >= 0),
    next_sync_at timestamptz NOT NULL,
    CHECK ((error_class IS NULL) = (failures = 0))
);

CREATE INDEX sync_attempts_instance ON ingest.sync_attempts (instance, ended_at DESC, id DESC);
CREATE INDEX sync_attempts_ended ON ingest.sync_attempts (ended_at DESC, id DESC);
CREATE INDEX import_batches_finished ON ingest.import_batches (finished_at DESC, id COLLATE "C" DESC)
    WHERE state = 'completed';

CREATE FUNCTION ingest.refuse_attempt_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'ingest.%: % refused, the sync history is append-only', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'OFI01';
END;
$$;

CREATE TRIGGER sync_attempts_append_only BEFORE UPDATE OR DELETE ON ingest.sync_attempts
    FOR EACH ROW EXECUTE FUNCTION ingest.refuse_attempt_change();
CREATE TRIGGER sync_attempts_no_truncate BEFORE TRUNCATE ON ingest.sync_attempts
    FOR EACH STATEMENT EXECUTE FUNCTION ingest.refuse_attempt_change();
