-- The valuation context's schema (docs/DESIGN.md D-20, D-21, D-25, D-36).
-- Only the valuation context reads or writes it. Amounts, prices, rates and
-- quantities are exact numerics, never floating point. Account IDs and
-- instruments refer to the ledger by value: no foreign key crosses schemas.

-- End-of-day prices (D-21), one per instrument (ISIN/symbol) and date.
CREATE TABLE valuation.prices (
    instrument text NOT NULL CHECK (instrument <> ''),
    date       date NOT NULL,
    close      numeric NOT NULL CHECK (close > 0),
    currency   char(3) NOT NULL,
    source     text NOT NULL,
    PRIMARY KEY (instrument, date)
);

-- FX rates to PLN (D-20), one per currency and date.
CREATE TABLE valuation.fx_rates (
    currency char(3) NOT NULL CHECK (currency <> 'PLN'),
    date     date NOT NULL,
    rate     numeric NOT NULL CHECK (rate > 0),
    source   text NOT NULL,
    PRIMARY KEY (currency, date)
);

-- Period index readings, one per kind and month, as yearly fractions.
CREATE TABLE valuation.index_readings (
    kind   text NOT NULL CHECK (kind IN ('cpi', 'reference_rate')),
    year   integer NOT NULL,
    month  integer NOT NULL CHECK (month BETWEEN 1 AND 12),
    value  numeric NOT NULL,
    source text NOT NULL,
    PRIMARY KEY (kind, year, month)
);

-- Series terms of retail treasury bonds, keyed by the series' ISIN. An
-- instrument whose ISIN is here is valued as a bond.
CREATE TABLE valuation.bond_terms (
    isin                 char(12) PRIMARY KEY,
    series               char(7) NOT NULL,
    first_period_rate    numeric NOT NULL CHECK (first_period_rate >= 0),
    margin               numeric NOT NULL,
    early_redemption_fee numeric NOT NULL CHECK (early_redemption_fee >= 0)
);

-- Each account's tax wrapper; an account without a row is taxable.
CREATE TABLE valuation.tax_wrappers (
    account_id text PRIMARY KEY CHECK (account_id <> ''),
    wrapper    text NOT NULL CHECK (wrapper IN ('taxable', 'ike', 'ikze'))
);

-- Stored valuations (D-36): one row per account, instrument, day and basis,
-- rebuilt by a revaluation. A row with a missing input has no value.
CREATE TABLE valuation.holding_valuations (
    account_id     text NOT NULL CHECK (account_id <> ''),
    isin           char(12) NOT NULL,
    symbol         text NOT NULL,
    date           date NOT NULL,
    basis          text NOT NULL CHECK (basis IN ('market', 'gross_accrued', 'net_if_redeemed')),
    quantity       numeric NOT NULL,
    value          numeric,
    value_currency char(3),
    pln            numeric,
    source         text NOT NULL,
    price_date     date,
    rate_date      date,
    missing        text CHECK (missing IN ('price', 'fx_rate', 'index_reading')),
    detail         text NOT NULL,
    PRIMARY KEY (account_id, isin, symbol, date, basis),
    CHECK ((missing IS NULL) = (value IS NOT NULL AND value_currency IS NOT NULL AND pln IS NOT NULL))
);

-- Each account's valued-through date: the last day whose holdings all have
-- a valuation.
CREATE TABLE valuation.account_status (
    account_id     text PRIMARY KEY CHECK (account_id <> ''),
    valued_through date NOT NULL,
    revalued_at    timestamptz NOT NULL DEFAULT now()
);

-- Pending revaluations, coalesced per account (D-36). from_date is the
-- earliest affected date not yet picked up: every enqueue lowers it with
-- LEAST(). A worker moves it to claimed_from when it starts and deletes the
-- row when it finishes with nothing new, so a change that arrives meanwhile
-- is kept, and a failed run is retried from claimed_from.
CREATE TABLE valuation.dirty (
    account_id   text PRIMARY KEY CHECK (account_id <> ''),
    from_date    date,
    claimed_from date,
    CHECK (from_date IS NOT NULL OR claimed_from IS NOT NULL)
);

-- Settings that, when changed, revalue everything: the bond basis (D-36).
CREATE TABLE valuation.settings (
    key   text PRIMARY KEY,
    value text NOT NULL
);
