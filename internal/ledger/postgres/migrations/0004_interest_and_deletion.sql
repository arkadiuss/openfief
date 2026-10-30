-- D-101: interest is a transaction kind of its own, such as a treasury bond's coupon.
-- D-100: a deletion is a version with no postings that deletes its key's transaction as a correction; the version it
-- supersedes stays (D-18).
ALTER TABLE ledger.transaction_versions DROP CONSTRAINT transaction_versions_kind_check;
ALTER TABLE ledger.transaction_versions ADD CONSTRAINT transaction_versions_kind_check CHECK (kind IN (
    'buy', 'sell', 'dividend', 'fee', 'fx-conversion', 'deposit', 'withdrawal', 'opening-lot', 'interest', 'deletion'));
