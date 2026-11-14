-- A delivery whose database transaction ran past its timeout (docs/DESIGN.md D-18, D-198) fails the sync attempt with
-- its own error class, delivery-timeout, so the history says it was the core's database that was too slow, not the
-- connector. The check names every class; the history is append-only, so this only widens it.
ALTER TABLE ingest.sync_attempts DROP CONSTRAINT sync_attempts_error_class_check;
ALTER TABLE ingest.sync_attempts ADD CONSTRAINT sync_attempts_error_class_check CHECK (error_class IN (
    'timeout', 'unavailable', 'connector-error', 'invalid-delivery', 'refused', 'booking', 'interrupted', 'other',
    'delivery-timeout'));
