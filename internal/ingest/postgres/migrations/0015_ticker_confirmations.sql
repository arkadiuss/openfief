-- D-153: a ticker's confirmed ISIN holds for every account at the same institution, and the owner can correct it,
-- which supersedes the transactions it booked. Confirmations move from each instance's ticker to the institution's,
-- one row per version: the highest is current, the earlier ones are kept. Each keeps what its listing needs
-- (ingest.ListingOf): the ISIN, the market's country, the trading currency of the market-data connector's proposal
-- when the owner confirmed that ISIN, and the owner's pick (D-170), '' for none.
ALTER TABLE ingest.tickers ADD COLUMN institution text NOT NULL DEFAULT '';
UPDATE ingest.tickers t SET institution = b.institution
    FROM (SELECT DISTINCT ON (instance) instance, institution FROM ingest.broker_accounts
          ORDER BY instance, institution) b
    WHERE b.instance = t.instance;
-- An instance that never reported a broker account is an institution of its own.
UPDATE ingest.tickers SET institution = 'instance:' || instance WHERE institution = '';
ALTER TABLE ingest.tickers ADD CONSTRAINT tickers_institution CHECK (institution <> '');

CREATE TABLE ingest.ticker_confirmations (
    institution       text NOT NULL CHECK (institution <> ''),
    symbol            text NOT NULL CHECK (symbol <> '' AND length(symbol) <= 32),
    version           integer NOT NULL CHECK (version >= 1),
    isin              text NOT NULL CHECK (isin ~ '^[A-Z]{2}[A-Z0-9]{9}[0-9]$'),
    country           text NOT NULL CHECK (country = '' OR country ~ '^[A-Z]{2}$'),
    proposal_currency text NOT NULL CHECK (length(proposal_currency) <= 3),
    picked            text NOT NULL CHECK (picked ~ '^([A-Za-z]{3})?$'),
    confirmed_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (institution, symbol, version)
);

-- The confirmations made so far, each instance's first per institution and symbol.
INSERT INTO ingest.ticker_confirmations (institution, symbol, version, isin, country, proposal_currency, picked, confirmed_at)
SELECT DISTINCT ON (institution, symbol) institution, symbol, 1, isin, country,
       CASE WHEN proposed_isin = isin THEN proposed_currency ELSE '' END, currency, confirmed_at
    FROM ingest.tickers WHERE isin IS NOT NULL
    ORDER BY institution, symbol, confirmed_at, instance;

ALTER TABLE ingest.tickers DROP COLUMN isin, DROP COLUMN currency, DROP COLUMN confirmed_at;

-- D-153: the entries a delivery booked are kept, marked booked, so a correction books them anew under the corrected
-- ISIN; those waiting for the owner are not booked yet. The entries the code before this migration deleted once it
-- booked them are no longer stored anywhere this schema can read, so no migration can bring them back (D-161): the
-- next delivery that names them, such as an upload of the same export, stores them again, marked booked, and until
-- then a correction books anew only the rows kept.
ALTER TABLE ingest.ticker_entries ADD COLUMN booked boolean NOT NULL DEFAULT false;

-- D-155: the holdings an institution reports for a ticker, such as the summed volume of its open positions in an
-- xStation export, asserted as balance assertions of the confirmed instrument, or waiting until the owner confirms
-- the ticker. One per account, ticker and date; the latest report replaces it. Quantities are exact decimal strings.
CREATE TABLE ingest.ticker_assertions (
    instance   text NOT NULL,
    symbol     text NOT NULL,
    account_id text NOT NULL CHECK (account_id <> ''),
    date       date NOT NULL,
    quantity   text NOT NULL,
    ref        text NOT NULL,
    asserted   boolean NOT NULL DEFAULT false,
    PRIMARY KEY (instance, symbol, account_id, date),
    FOREIGN KEY (instance, symbol) REFERENCES ingest.tickers (instance, symbol)
);
