-- The ledger context's schema (docs/DESIGN.md D-14, D-18, D-19). Only the
-- ledger reads or writes it; other contexts go through internal/ledger/app.
-- Amounts and quantities are exact numerics, never floating point.

-- Sleeves belong to the portfolio context, which has no store of its own in
-- M0; they live here next to the accounts that refer to them.
CREATE TABLE ledger.sleeves (
    id   text PRIMARY KEY CHECK (id <> ''),
    name text NOT NULL CHECK (name <> ''),
    role text NOT NULL CHECK (role IN ('defensive', 'offensive'))
);

CREATE TABLE ledger.accounts (
    id          text PRIMARY KEY CHECK (id <> ''),
    institution text NOT NULL CHECK (institution <> ''),
    name        text NOT NULL CHECK (name <> ''),
    sleeve_id   text NOT NULL REFERENCES ledger.sleeves (id)
);

-- One row per version of a transaction. A key (source, external_id) has
-- versions 1..n; the highest is current and the others are kept (D-18).
CREATE TABLE ledger.transaction_versions (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source      text NOT NULL CHECK (source <> ''),
    external_id text NOT NULL CHECK (external_id <> ''),
    version     integer NOT NULL CHECK (version >= 1),
    account_id  text NOT NULL REFERENCES ledger.accounts (id),
    kind        text NOT NULL CHECK (kind IN (
        'buy', 'sell', 'dividend', 'fee', 'fx-conversion', 'deposit', 'withdrawal', 'opening-lot')),
    date        date NOT NULL,
    lot_opened  date,
    description text NOT NULL,
    -- ledger.Fingerprint of the version's content, checked on every load.
    fingerprint char(64) NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT transaction_versions_key_version UNIQUE (source, external_id, version),
    CHECK ((kind = 'opening-lot') = (lot_opened IS NOT NULL))
);

CREATE INDEX transaction_versions_account ON ledger.transaction_versions (account_id);

-- One row per posting of a version, in the order it was given.
CREATE TABLE ledger.postings (
    version_id bigint NOT NULL REFERENCES ledger.transaction_versions (id),
    position   integer NOT NULL CHECK (position >= 0),
    book_kind  text NOT NULL CHECK (book_kind IN ('cash', 'holding', 'fee', 'withholding-tax', 'fx', 'equity')),
    -- The instrument of a holding; NULL for every other ledger account.
    isin       char(12),
    symbol     text,
    amount     numeric NOT NULL,
    -- The amount's currency, which is also the ledger account's currency for
    -- every kind but a holding.
    currency   char(3) NOT NULL,
    quantity   numeric NOT NULL DEFAULT 0,
    PRIMARY KEY (version_id, position),
    CHECK ((book_kind = 'holding') = (isin IS NOT NULL AND symbol IS NOT NULL)),
    CHECK (book_kind = 'holding' OR quantity = 0)
);

-- The log is append-only (D-18). It is enforced with triggers, not REVOKE:
-- in M0 the core connects as the role that owns the schema, and an owner
-- keeps every privilege a REVOKE would take away, while a trigger binds every
-- role, the owner included. Only a superuser can switch it off.
CREATE FUNCTION ledger.refuse_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'ledger.%: % refused, the transaction log is append-only', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'OFA01';
END;
$$;

CREATE TRIGGER transaction_versions_append_only
    BEFORE UPDATE OR DELETE ON ledger.transaction_versions
    FOR EACH ROW EXECUTE FUNCTION ledger.refuse_change();
CREATE TRIGGER transaction_versions_no_truncate
    BEFORE TRUNCATE ON ledger.transaction_versions
    FOR EACH STATEMENT EXECUTE FUNCTION ledger.refuse_change();
CREATE TRIGGER postings_append_only
    BEFORE UPDATE OR DELETE ON ledger.postings
    FOR EACH ROW EXECUTE FUNCTION ledger.refuse_change();
CREATE TRIGGER postings_no_truncate
    BEFORE TRUNCATE ON ledger.postings
    FOR EACH STATEMENT EXECUTE FUNCTION ledger.refuse_change();

-- A version must be the next one of its key: version n needs version n-1.
CREATE FUNCTION ledger.require_previous_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.version > 1 AND NOT EXISTS (
        SELECT 1 FROM ledger.transaction_versions
        WHERE source = NEW.source AND external_id = NEW.external_id AND version = NEW.version - 1
    ) THEN
        RAISE EXCEPTION 'ledger.transaction_versions: version % of %/% has no version %',
            NEW.version, NEW.source, NEW.external_id, NEW.version - 1
            USING ERRCODE = 'OFA02';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER transaction_versions_in_sequence
    BEFORE INSERT ON ledger.transaction_versions
    FOR EACH ROW EXECUTE FUNCTION ledger.require_previous_version();

-- Postings are written right after their version, in the same database
-- transaction, and never added to a version stored earlier. Inserting a
-- version opens it for postings until the transaction ends or the next
-- version is inserted. (A row's xmin can't tell this: Store.Append runs in a
-- savepoint inside an import, which has its own transaction ID.)
CREATE FUNCTION ledger.open_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM set_config('ledger.open_version', NEW.id::text, true);
    RETURN NEW;
END;
$$;

CREATE TRIGGER transaction_versions_open
    AFTER INSERT ON ledger.transaction_versions
    FOR EACH ROW EXECUTE FUNCTION ledger.open_version();

CREATE FUNCTION ledger.require_open_version() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF coalesce(current_setting('ledger.open_version', true), '') <> NEW.version_id::text THEN
        RAISE EXCEPTION 'ledger.postings: version row % is not the one just inserted', NEW.version_id
            USING ERRCODE = 'OFA01';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER postings_with_their_version
    BEFORE INSERT ON ledger.postings
    FOR EACH ROW EXECUTE FUNCTION ledger.require_open_version();

-- The latest check of each balance assertion. A new report or a new check of
-- the same account, date and ledger account replaces the row; only the
-- transaction log is append-only.
CREATE TABLE ledger.assertion_checks (
    account_id text NOT NULL REFERENCES ledger.accounts (id),
    date       date NOT NULL,
    -- The ledger account, "cash:<currency>" or "holding:<ISIN>/<symbol>".
    book       text NOT NULL,
    book_kind  text NOT NULL CHECK (book_kind IN ('cash', 'holding')),
    currency   char(3),
    isin       char(12),
    symbol     text,
    balance    numeric NOT NULL,
    status     text NOT NULL CHECK (status IN ('matched', 'flagged')),
    derived    numeric NOT NULL,
    checked_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, date, book),
    CHECK (CASE book_kind
        WHEN 'cash' THEN currency IS NOT NULL AND isin IS NULL AND symbol IS NULL
        ELSE currency IS NULL AND isin IS NOT NULL AND symbol IS NOT NULL
    END)
);
