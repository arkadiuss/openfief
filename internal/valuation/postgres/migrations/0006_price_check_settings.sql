-- The owner's price check thresholds (docs/DESIGN.md D-110): a cross-check price is flagged when it differs from the
-- stored price by more than 1% of it, unless the owner set another threshold for its instrument, such as a thinly
-- traded Warsaw listing. One row per instrument; an instrument without one uses 1%.
CREATE TABLE valuation.price_check_thresholds (
    instrument text NOT NULL PRIMARY KEY CHECK (instrument <> ''),
    threshold  numeric NOT NULL CHECK (threshold > 0 AND threshold < 1),
    set_at     timestamptz NOT NULL DEFAULT now()
);

-- The owner's acknowledgements of flagged price checks (D-112), recorded append-only: the check's instrument, the
-- cross-check price's date, source and close, and the date and close of the price it was compared with. An
-- acknowledgement covers the check only while both prices stay the same.
CREATE TABLE valuation.price_check_acknowledgements (
    id                   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    instrument           text NOT NULL CHECK (instrument <> ''),
    date                 date NOT NULL,
    source               text NOT NULL CHECK (source <> ''),
    cross_check          numeric NOT NULL CHECK (cross_check > 0),
    cross_check_currency char(3) NOT NULL,
    price_date           date NOT NULL CHECK (price_date <= date),
    price                numeric NOT NULL CHECK (price > 0),
    price_currency       char(3) NOT NULL,
    acknowledged_at      timestamptz NOT NULL DEFAULT now()
);

-- Append-only, enforced with triggers as the ledger's log is (ledger/postgres 0001): a trigger binds every role, the
-- schema's owner included.
CREATE FUNCTION valuation.refuse_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'valuation.%: % refused, it is append-only', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'OFA01';
END;
$$;

CREATE TRIGGER price_check_acknowledgements_append_only
    BEFORE UPDATE OR DELETE ON valuation.price_check_acknowledgements
    FOR EACH ROW EXECUTE FUNCTION valuation.refuse_change();
CREATE TRIGGER price_check_acknowledgements_no_truncate
    BEFORE TRUNCATE ON valuation.price_check_acknowledgements
    FOR EACH STATEMENT EXECUTE FUNCTION valuation.refuse_change();
