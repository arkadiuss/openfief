-- Ranked market-data sources (docs/DESIGN.md D-94): every source's prices, FX rates and index readings are stored,
-- one per key and source, and the owner ranks the sources of each kind. For each key the highest-ranked source with a
-- value wins; the lower-ranked ones serve as cross-checks (D-22). A source is the connector instance that delivered
-- the value.
ALTER TABLE valuation.prices DROP CONSTRAINT prices_pkey, ADD PRIMARY KEY (instrument, date, source);
ALTER TABLE valuation.fx_rates DROP CONSTRAINT fx_rates_pkey, ADD PRIMARY KEY (currency, date, source);
ALTER TABLE valuation.index_readings DROP CONSTRAINT index_readings_pkey, ADD PRIMARY KEY (kind, year, month, source);

-- Every source that delivered a kind, with the institution it named, which places it by default: the first-party
-- connector of each kind ranks first unless the owner ranks otherwise.
CREATE TABLE valuation.sources (
    kind        text NOT NULL CHECK (kind IN ('prices', 'fx_rates', 'index_readings')),
    source      text NOT NULL CHECK (source <> '' AND length(source) <= 64),
    institution text NOT NULL,
    PRIMARY KEY (kind, source)
);

-- The owner's ranking of each kind, highest first: position 0 ranks first. A source not listed ranks after every
-- listed one, in its default place.
CREATE TABLE valuation.source_rankings (
    kind     text NOT NULL CHECK (kind IN ('prices', 'fx_rates', 'index_readings')),
    position integer NOT NULL CHECK (position >= 0),
    source   text NOT NULL CHECK (source <> '' AND length(source) <= 64),
    PRIMARY KEY (kind, position),
    UNIQUE (kind, source)
);
