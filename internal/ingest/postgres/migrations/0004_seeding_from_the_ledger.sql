-- Whether an instance syncs in opening or transactions mode now follows the ledger: each bound account is seeded or
-- not by its own opening lots and opening cash (docs/DESIGN.md D-23, D-95), so the instance no longer keeps a seeded
-- flag of its own.
ALTER TABLE ingest.connector_instances DROP COLUMN seeded;
