-- D-135, D-162: a hand-entered bond account's payout setting is an account setting connectors read, so the ingest
-- context keeps it (docs/dev/code-organization.md), keyed by the ledger's account ID; the ledger knows nothing of it.
-- payout is the setting in force, payout_since the day it took effect (NULL for an account's first setting, which
-- covers its whole history), and payout_earlier the settings in force before it, oldest first, as a JSON array of
-- {"payout": "cash" | "paid-out", "since": "YYYY-MM-DD" or ""}. An account without a row keeps its payouts as cash.
-- Nothing is copied from the ledger schema's old columns: what was stored before the first release is entered again
-- (D-161).
CREATE TABLE ingest.payout_settings (
    account_id     text PRIMARY KEY CHECK (account_id <> ''),
    payout         text NOT NULL CHECK (payout IN ('cash', 'paid-out')),
    payout_since   date,
    payout_earlier jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(payout_earlier) = 'array')
);
