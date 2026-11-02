-- An account's form, which the owner sets when creating an OKI account (docs/DESIGN.md D-147): a key of the
-- jurisdiction pack's, such as bank-account or brokerage-account. It decides whether the account's cash is a savings or
-- an investment asset for OKI's yearly tax, which is computed when read, so it changes no stored valuation.
CREATE TABLE valuation.account_forms (
    account_id text PRIMARY KEY CHECK (account_id <> ''),
    form       text NOT NULL CHECK (form <> ''),
    set_at     timestamptz NOT NULL DEFAULT now()
);

-- The total an IKE's owner paid into it before its history is known, entered for an account seeded from a starting
-- balance (D-23, D-149): its early return is taxed above it. Totals are taxed when read, so it changes no stored
-- valuation either.
CREATE TABLE valuation.earlier_payments (
    account_id text PRIMARY KEY CHECK (account_id <> ''),
    amount     numeric NOT NULL CHECK (amount >= 0),
    currency   char(3) NOT NULL,
    set_at     timestamptz NOT NULL DEFAULT now()
);
