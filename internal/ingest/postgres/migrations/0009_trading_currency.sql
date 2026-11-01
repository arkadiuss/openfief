-- D-159: an XTB open position's open price is in its instrument's trading currency, which the export doesn't name: the
-- currency of the confirmed ticker's listing, or of the exchange its market names. A proposal keeps the currency its
-- listing trades in, as the market-data connector names it ("GBX" for pence), and an opening lot priced that way waits
-- with its open price and no amount or currency until the owner confirms its ticker.
ALTER TABLE ingest.tickers ADD COLUMN proposed_currency text NOT NULL DEFAULT ''
    CHECK (proposed_currency ~ '^([A-Za-z]{3})?$');

ALTER TABLE ingest.ticker_entries ADD COLUMN open_price text;
ALTER TABLE ingest.ticker_entries DROP CONSTRAINT ticker_entries_currency_check;
ALTER TABLE ingest.ticker_entries ADD CONSTRAINT ticker_entries_currency_check
    CHECK (currency ~ '^[A-Z]{3}$' OR (currency = '' AND kind = 'opening-lot' AND open_price IS NOT NULL));
