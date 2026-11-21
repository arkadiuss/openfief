-- The thesis context's schema: the owner's draft theses and every version of each (docs/DESIGN.md D-229). Only the thesis
-- context reads or writes it.

-- One row per version of a draft thesis: version 1 is the save, a later one an edit. A thesis is its rows; its latest
-- version is the highest number, and the earlier ones stay (D-229). It has no agent, no fief and no lease, and no column
-- for a lease's size: the control plane starts a fief from a draft later, as a proposal the owner approves (GR-6).
-- The text is T1 data (D-206).
CREATE TABLE thesis.thesis (
    thesis_id    text        NOT NULL CHECK (thesis_id ~ '^[0-9a-f]{16}$'),
    version      integer     NOT NULL CHECK (version >= 1),
    saved_at     timestamptz NOT NULL,
    statement    text        NOT NULL CHECK (statement <> '' AND char_length(statement) <= 1000),
    horizon      text        NOT NULL CHECK (horizon IN ('long_term', 'short_term')),
    instruments  text[]      NOT NULL CHECK (cardinality(instruments) BETWEEN 1 AND 10),
    invalidation text[]      NOT NULL CHECK (cardinality(invalidation) BETWEEN 1 AND 8),
    sources      text[]      NOT NULL CHECK (cardinality(sources) BETWEEN 1 AND 10),
    PRIMARY KEY (thesis_id, version)
);

-- The list reads the newest change first.
CREATE INDEX thesis_saved_at ON thesis.thesis (saved_at DESC);

-- A saved version is never changed or removed: an edit adds a version. A statement trigger fires even when the statement
-- matches no row. Like the audit log's, this is no defence against the database's own administrator.
CREATE FUNCTION thesis.thesis_refuse() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'thesis.thesis keeps every version: % refused', TG_OP USING ERRCODE = 'restrict_violation';
END
$$;

CREATE TRIGGER thesis_keeps_versions
    BEFORE UPDATE OR DELETE OR TRUNCATE ON thesis.thesis
    FOR EACH STATEMENT EXECUTE FUNCTION thesis.thesis_refuse();
