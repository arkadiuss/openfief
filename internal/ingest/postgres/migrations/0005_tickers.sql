-- Tickers without an ISIN (docs/DESIGN.md D-116): the instruments a connector instance named by ticker only, such as
-- XTB's, the ISIN a market-data connector proposed for each, the owner's confirmation, and the entries waiting for it.
CREATE TABLE ingest.tickers (
    instance      text NOT NULL CHECK (instance ~ '^[a-z]([-a-z0-9]*[a-z0-9])?$' AND length(instance) <= 56),
    symbol        text NOT NULL CHECK (symbol <> '' AND symbol !~ '[[:space:][:cntrl:]]' AND length(symbol) <= 32),
    code          text NOT NULL CHECK (length(code) <= 32),
    country       text NOT NULL CHECK (country = '' OR country ~ '^[A-Z]{2}$'),
    name          text NOT NULL CHECK (length(name) <= 200),
    proposed_isin text CHECK (proposed_isin ~ '^[A-Z]{2}[A-Z0-9]{9}[0-9]$'),
    proposed_name text NOT NULL DEFAULT '' CHECK (length(proposed_name) <= 200),
    -- Set once by the owner (D-116) and never changed: every entry of the ticker is booked with it.
    isin          text CHECK (isin ~ '^[A-Z]{2}[A-Z0-9]{9}[0-9]$'),
    confirmed_at  timestamptz,
    first_seen    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (instance, symbol),
    CHECK ((isin IS NULL) = (confirmed_at IS NULL))
);

-- The buys, sells and dividends of a ticker not confirmed yet, keyed as the delivery keyed them (source without the
-- instance's prefix), until the owner's confirmation books them. Amounts are exact decimal strings.
CREATE TABLE ingest.ticker_entries (
    instance        text NOT NULL,
    key_source      text NOT NULL CHECK (key_source <> ''),
    external_id     text NOT NULL CHECK (external_id <> ''),
    symbol          text NOT NULL,
    account_id      text NOT NULL CHECK (account_id <> ''),
    kind            text NOT NULL CHECK (kind IN ('buy', 'sell', 'dividend')),
    date            date NOT NULL,
    quantity        text NOT NULL,
    amount          text NOT NULL,
    currency        text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    charge          text NOT NULL,
    charge_currency text NOT NULL CHECK (charge_currency = '' OR charge_currency ~ '^[A-Z]{3}$'),
    description     text NOT NULL,
    ref             text NOT NULL,
    PRIMARY KEY (instance, key_source, external_id),
    FOREIGN KEY (instance, symbol) REFERENCES ingest.tickers (instance, symbol)
);
