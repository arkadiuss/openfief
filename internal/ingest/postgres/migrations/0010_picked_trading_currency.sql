-- D-170: when the confirmed listing names no trading currency and the exchange the ticker's market suffix names quotes
-- in more than one, as London does in pence (GBX), pounds and dollars, the owner picks it when confirming the ticker's
-- ISIN, and no currency is assumed. The pick is kept with the confirmation; '' is none, which every ticker confirmed
-- before this migration has (D-161: nothing stored before the first release is migrated).
ALTER TABLE ingest.tickers ADD COLUMN currency text NOT NULL DEFAULT ''
    CHECK (currency ~ '^([A-Za-z]{3})?$');
