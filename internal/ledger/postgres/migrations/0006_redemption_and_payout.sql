-- D-135: a treasury bond redeemed at maturity is a transaction kind of its own, and each hand-entered bond account has
-- a payout setting: coupons and maturities stay as the account's cash, or are paid out. payout_since is the day the
-- setting took effect, so a change never rewrites what fell due before it.
ALTER TABLE ledger.transaction_versions DROP CONSTRAINT transaction_versions_kind_check;
ALTER TABLE ledger.transaction_versions ADD CONSTRAINT transaction_versions_kind_check CHECK (kind IN (
    'buy', 'sell', 'dividend', 'fee', 'fx-conversion', 'deposit', 'withdrawal', 'opening-lot', 'interest', 'deletion',
    'redemption'));

ALTER TABLE ledger.accounts
    ADD COLUMN payout text NOT NULL DEFAULT 'cash' CHECK (payout IN ('cash', 'paid-out')),
    ADD COLUMN payout_since date,
    ADD CONSTRAINT accounts_paid_out_since CHECK (payout = 'cash' OR payout_since IS NOT NULL);
