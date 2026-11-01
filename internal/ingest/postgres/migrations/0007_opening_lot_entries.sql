-- An XTB account may start from the open positions of a current export (docs/DESIGN.md D-136): each position is an
-- opening lot named by ticker only, which waits for its ticker's ISIN like a buy does (D-116), with the date its units
-- were first acquired.
ALTER TABLE ingest.ticker_entries DROP CONSTRAINT ticker_entries_kind_check;
ALTER TABLE ingest.ticker_entries ADD CONSTRAINT ticker_entries_kind_check
    CHECK (kind IN ('buy', 'sell', 'dividend', 'opening-lot'));
ALTER TABLE ingest.ticker_entries ADD COLUMN lot_opened date;
ALTER TABLE ingest.ticker_entries ADD CONSTRAINT ticker_entries_lot_opened_check
    CHECK ((kind = 'opening-lot') = (lot_opened IS NOT NULL));
