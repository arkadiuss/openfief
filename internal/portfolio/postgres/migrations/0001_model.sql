-- The portfolio context's schema: the optional model portfolio (docs/DESIGN.md
-- D-52). Only the portfolio context reads or writes it. Sleeves and sleeve
-- kinds stay in the ledger schema, next to the accounts that refer to them.

-- Whether the model portfolio is on. At most one row; no row means it was
-- never saved and is off (D-52: off by default).
CREATE TABLE portfolio.model (
    singleton  boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    enabled    boolean NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- The model's targets in the owner's order. An instrument is named
-- "ISIN/symbol" and has no currency; an index is named as the market-data
-- connector names it and has the currency of its levels. Weights are in
-- percent with at most two decimal places; that they sum to 100 is checked
-- by the portfolio context before saving.
CREATE TABLE portfolio.model_targets (
    position integer PRIMARY KEY CHECK (position >= 0),
    kind     text NOT NULL CHECK (kind IN ('instrument', 'index')),
    name     text NOT NULL CHECK (name <> '' AND length(name) <= 64),
    currency char(3),
    weight   numeric NOT NULL CHECK (weight > 0 AND weight <= 100 AND scale(weight) <= 2),
    UNIQUE (kind, name),
    CHECK ((kind = 'index') = (currency IS NOT NULL))
);
