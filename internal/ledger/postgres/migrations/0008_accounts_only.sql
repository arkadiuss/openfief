-- The ledger knows nothing of sleeves: the portfolio context keeps the sleeve kinds, the sleeves and which account is
-- in which sleeve in its own schema, and the ingest context keeps a bond account's payout setting (D-135, D-162) in
-- its own (docs/dev/code-organization.md). Nothing is copied over: what was stored before the first release is
-- entered again, and the demo reseeded (D-161).
ALTER TABLE ledger.accounts
    DROP COLUMN sleeve_id,
    DROP COLUMN payout,
    DROP COLUMN payout_since,
    DROP COLUMN payout_earlier;

DROP TABLE ledger.sleeves;
DROP TABLE ledger.sleeve_kinds;
DROP FUNCTION ledger.sleeve_role_override();
DROP FUNCTION ledger.sleeve_kind_role_fixed();
