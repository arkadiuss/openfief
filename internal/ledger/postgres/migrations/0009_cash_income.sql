-- D-152: cash income is a transaction kind of its own, for income that no holding produces, such as a broker's
-- interest on uninvested cash or a closed CFD's result, posted against a ledger account of its own kind, income.
ALTER TABLE ledger.transaction_versions DROP CONSTRAINT transaction_versions_kind_check;
ALTER TABLE ledger.transaction_versions ADD CONSTRAINT transaction_versions_kind_check CHECK (kind IN (
    'buy', 'sell', 'dividend', 'fee', 'fx-conversion', 'deposit', 'withdrawal', 'opening-lot', 'interest', 'deletion',
    'redemption', 'cash-income'));

ALTER TABLE ledger.postings DROP CONSTRAINT postings_book_kind_check;
ALTER TABLE ledger.postings ADD CONSTRAINT postings_book_kind_check CHECK (book_kind IN (
    'cash', 'holding', 'fee', 'withholding-tax', 'fx', 'equity', 'income'));
