-- The OpenFief accounts a completed import batch booked into (docs/DESIGN.md D-109, D-140): every mapped account a
-- statement of several accounts books, so the recent activity names them and shows each batch only to callers who may
-- view one of them (D-163, D-193). Internal account IDs only, never a broker account number (GR-12). NULL for batches
-- completed before this column existed and for open ones; an empty array for a batch that booked into none.
ALTER TABLE ingest.import_batches ADD COLUMN accounts text[];
