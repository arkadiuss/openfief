-- Totals count tax as PIT-38 does (docs/DESIGN.md D-128): a holding priced at market keeps, with its net value, its
-- cost basis for tax and what selling it today would cost, in PLN, so a total can offset its gain against other
-- holdings' losses. Rows stored before have neither; the pack's rules version rose with this change, so the core
-- revalues every account at start-up (app.InputRecorder.ApplyRules).
ALTER TABLE valuation.holding_valuations
    ADD COLUMN tax_cost numeric,
    ADD COLUMN sale_cost numeric,
    ADD CONSTRAINT holding_valuations_tax_cost_check
        CHECK ((tax_cost IS NULL) = (sale_cost IS NULL) AND (tax_cost IS NULL OR (basis = 'market' AND net IS NOT NULL)));

-- The owner's classification of an instrument for OKI's exemptions (D-127, D-132), for an instrument the
-- jurisdiction pack can't classify: one row per instrument, the class a key of the pack's (savings, investment or
-- other). It stays with the instrument; setting it again replaces it. It changes no stored valuation: the yearly tax
-- is computed when read.
CREATE TABLE valuation.exemption_classes (
    instrument text NOT NULL PRIMARY KEY CHECK (instrument <> ''),
    class      text NOT NULL CHECK (class <> ''),
    set_at     timestamptz NOT NULL DEFAULT now()
);
