-- End-of-day closing levels of market indices, such as WIG20 (D-51), one
-- per index and date, named as the market-data connector names the index.
-- The model portfolio's index targets read them (D-52, D-53); no holding is
-- valued from them, so they enqueue no revaluation.
CREATE TABLE valuation.index_levels (
    index  text NOT NULL CHECK (index <> '' AND length(index) <= 64),
    date   date NOT NULL,
    level  numeric NOT NULL CHECK (level > 0),
    source text NOT NULL,
    PRIMARY KEY (index, date)
);
