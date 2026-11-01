-- D-162: an account's first payout setting covers its whole history, so a paid-out setting may have no day it took
-- effect. A change takes effect from a day the owner gives, possibly earlier than today, and what fell due before it
-- follows the setting in force then: payout_earlier keeps those earlier settings, oldest first, as a JSON array of
-- {"payout": "cash" | "paid-out", "since": "YYYY-MM-DD" or ""}.
ALTER TABLE ledger.accounts
    DROP CONSTRAINT accounts_paid_out_since,
    ADD COLUMN payout_earlier jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(payout_earlier) = 'array');
