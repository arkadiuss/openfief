-- D-158: the owner enters the cash an institution shows for an account on a date, a cash checkpoint. It is stored as
-- a balance assertion like a connector's report; entered tells the two apart, so its mismatch raises its own
-- attention item and the owner can book the difference from it. Only cash is entered.
ALTER TABLE ledger.assertion_checks
    ADD COLUMN entered boolean NOT NULL DEFAULT false,
    ADD CONSTRAINT assertion_checks_entered_cash CHECK (NOT entered OR book_kind = 'cash');
