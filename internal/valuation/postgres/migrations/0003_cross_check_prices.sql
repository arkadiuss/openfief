-- Cross-check prices (D-22): institutions' own end-of-day prices, such as
-- IBKR's mark prices, kept beside valuation.prices to cross-check the
-- market-data connector's closes. No holding is valued from them, so they
-- enqueue no revaluation. One per instrument, date and source, so two
-- connector instances that report the same instrument both keep theirs.
CREATE TABLE valuation.cross_check_prices (
    instrument text NOT NULL CHECK (instrument <> ''),
    date       date NOT NULL,
    close      numeric NOT NULL CHECK (close > 0),
    currency   char(3) NOT NULL,
    source     text NOT NULL CHECK (source <> ''),
    PRIMARY KEY (instrument, date, source)
);
