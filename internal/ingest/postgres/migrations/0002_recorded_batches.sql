-- Import batches the owner started (docs/DESIGN.md D-49): a statement file uploaded in the built-in UI ('upload') and
-- a manual entry such as a treasury bond lot ('entry') are recorded in the batch history next to syncs, once they
-- complete. Still only IDs, counts and times (GR-12).
ALTER TABLE ingest.import_batches DROP CONSTRAINT import_batches_source_check;
ALTER TABLE ingest.import_batches ADD CONSTRAINT import_batches_source_check
    CHECK (source IN ('sync', 'file', 'upload', 'entry'));
