-- D-197: an instrument has a display name, the first one a connector delivers for it, which the owner can change.
-- Names are kept per ISIN and append-only like the rest of the ledger (D-18): a rename is a new row. The first row
-- a connector delivers for an ISIN stays the delivered name (one per ISIN, the partial unique index); the owner's
-- latest row wins over it (ledger.CurrentInstrumentNames). source names the connector instance a delivered name came
-- from, and is empty for the owner's.
CREATE TABLE ledger.instrument_names (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    isin        text NOT NULL CHECK (isin ~ '^[A-Z]{2}[A-Z0-9]{9}[0-9]$'),
    name        text NOT NULL CHECK (name <> '' AND char_length(name) <= 200),
    by_owner    boolean NOT NULL,
    source      text NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    CHECK (by_owner = (source = ''))
);

CREATE UNIQUE INDEX instrument_names_first_delivered ON ledger.instrument_names (isin) WHERE NOT by_owner;

CREATE TRIGGER instrument_names_append_only
    BEFORE UPDATE OR DELETE ON ledger.instrument_names
    FOR EACH ROW EXECUTE FUNCTION ledger.refuse_change();
CREATE TRIGGER instrument_names_no_truncate
    BEFORE TRUNCATE ON ledger.instrument_names
    FOR EACH STATEMENT EXECUTE FUNCTION ledger.refuse_change();
