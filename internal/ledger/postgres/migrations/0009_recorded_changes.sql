-- The read API's recent activity (docs/DESIGN.md D-193) lists the coupons booked and the corrections recorded, newest
-- first, a page at a time (Store.RecordedChanges): every version after a transaction's first, deletions included, and
-- the first version of an interest transaction. This index serves that order without reading the whole log.
CREATE INDEX transaction_versions_recorded_changes ON ledger.transaction_versions (recorded_at DESC, id DESC)
    WHERE version > 1 OR kind = 'interest';
