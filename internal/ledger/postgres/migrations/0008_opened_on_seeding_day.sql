-- D-168: the owner can mark an account as opened on its seeding day, so its seeding from a starting balance doesn't
-- move the portfolio's charts' default start (D-138). Accounts stored before are unmarked: their seeding still moves it.
ALTER TABLE ledger.accounts
    ADD COLUMN opened_on_seeding_day boolean NOT NULL DEFAULT false;
