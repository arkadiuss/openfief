-- D-160: an opening upload starts its account at once. Its go-live is recorded with the upload, even when nothing can
-- be booked yet, such as positions that wait for their ticker's ISIN (D-116): later uploads are then transactions, and
-- the waiting positions join the opening on this date. The ledger's own go-live, from its seeding transactions, wins
-- once there is one. The first opening upload of an account sets it; nothing changes it after.
CREATE TABLE ingest.account_starts (
    account_id  text PRIMARY KEY CHECK (account_id <> ''),
    go_live     date NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now()
);
