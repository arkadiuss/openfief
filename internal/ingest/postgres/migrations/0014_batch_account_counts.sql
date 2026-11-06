-- A completed import batch's counts per account it booked into (docs/DESIGN.md D-199): the transactions its mapping
-- produced there, what the ledger did with them and the balance assertions that didn't match, so a caller who may view
-- only some of a batch's accounts sees only theirs. A JSON array of {"account", "transactions", "new", "newVersions",
-- "unchanged", "failedAssertions"} in account order; internal account IDs and counts only, never an amount or a broker
-- account number (GR-12). NULL for batches completed before this column existed and for open ones.
ALTER TABLE ingest.import_batches ADD COLUMN account_counts jsonb;
