-- Attention items (docs/DESIGN.md D-130): one item per kind and subject, its history of changes, and the core's
-- outbox, whose events publish each change. Everything is append-only, enforced with triggers as in the ledger
-- schema: in M0 the core connects as the role that owns the schema, and a trigger binds every role, the owner
-- included. A change and its outbox event are written in one database transaction (internal/attention/postgres).
CREATE TABLE attention.items (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind       text NOT NULL CHECK (kind ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$'),
    subject    text NOT NULL CHECK (subject <> '' AND octet_length(subject) <= 300),
    created_at timestamptz NOT NULL,
    UNIQUE (kind, subject)
);

CREATE TABLE attention.changes (
    item_id bigint NOT NULL REFERENCES attention.items (id),
    -- Numbers the item's changes from 1, without gaps.
    seq     integer NOT NULL CHECK (seq >= 1),
    state   text NOT NULL CHECK (state IN ('open', 'acknowledged', 'resolved')),
    reason  text NOT NULL CHECK (reason IN ('raised', 'changed', 'reopened', 'cleared', 'acknowledged', 'resolved')),
    detail  text NOT NULL CHECK (octet_length(detail) <= 500),
    at      timestamptz NOT NULL,
    PRIMARY KEY (item_id, seq)
);

-- The outbox: one event per change, published in id order. payload is the event as
-- openfief.attention.v1.AttentionItemChanged's JSON form, without event_id, which is the row's id.
CREATE TABLE attention.outbox (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic      text NOT NULL CHECK (topic <> ''),
    item_id    bigint NOT NULL,
    seq        integer NOT NULL,
    payload    jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    FOREIGN KEY (item_id, seq) REFERENCES attention.changes (item_id, seq),
    UNIQUE (item_id, seq)
);

-- Publishing an event is recorded by a row here, never by changing the event.
CREATE TABLE attention.published (
    event_id     bigint PRIMARY KEY REFERENCES attention.outbox (id),
    published_at timestamptz NOT NULL
);

CREATE FUNCTION attention.refuse_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'attention.%: % refused, attention items and their outbox are append-only', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'OFA01';
END;
$$;

CREATE TRIGGER items_append_only BEFORE UPDATE OR DELETE ON attention.items
    FOR EACH ROW EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER items_no_truncate BEFORE TRUNCATE ON attention.items
    FOR EACH STATEMENT EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER changes_append_only BEFORE UPDATE OR DELETE ON attention.changes
    FOR EACH ROW EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER changes_no_truncate BEFORE TRUNCATE ON attention.changes
    FOR EACH STATEMENT EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER outbox_append_only BEFORE UPDATE OR DELETE ON attention.outbox
    FOR EACH ROW EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER outbox_no_truncate BEFORE TRUNCATE ON attention.outbox
    FOR EACH STATEMENT EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER published_append_only BEFORE UPDATE OR DELETE ON attention.published
    FOR EACH ROW EXECUTE FUNCTION attention.refuse_change();
CREATE TRIGGER published_no_truncate BEFORE TRUNCATE ON attention.published
    FOR EACH STATEMENT EXECUTE FUNCTION attention.refuse_change();

-- A change continues its item's history: seq 1 for a new item, one more than the latest otherwise.
CREATE FUNCTION attention.require_next_seq() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.seq <> COALESCE((SELECT max(seq) FROM attention.changes WHERE item_id = NEW.item_id), 0) + 1 THEN
        RAISE EXCEPTION 'attention.changes: change % of item % is out of sequence', NEW.seq, NEW.item_id
            USING ERRCODE = 'OFA02';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER changes_in_sequence BEFORE INSERT ON attention.changes
    FOR EACH ROW EXECUTE FUNCTION attention.require_next_seq();
