-- One source per market-data series (docs/DESIGN.md D-94, rewritten): each instrument's prices, each currency's FX
-- rates and each period index's readings take every value from one connector instance. It is chosen from the kind's
-- ranking (0004) when the series first needs a value, or by the owner, and recorded here; every other source's
-- values stay stored as cross-checks. Series already held when this migration runs get theirs the same way, on the
-- first revaluation after it, which the new rules version (valuation.SourceRulesVersion) enqueues for every account.
-- Market index levels are no series and have no row (D-124).
CREATE TABLE valuation.series_sources (
    kind      text NOT NULL CHECK (kind IN ('prices', 'fx_rates', 'index_readings')),
    key       text NOT NULL CHECK (key <> '' AND length(key) <= 200),
    source    text NOT NULL CHECK (source <> '' AND length(source) <= 64),
    by_owner  boolean NOT NULL DEFAULT false,
    chosen_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (kind, key)
);

-- The owner's gap threshold of each kind, in days: a series whose source's latest value is older than that on the
-- day valued raises an attention item (D-94, D-130). A kind without a row uses valuation.DefaultGapThreshold.
CREATE TABLE valuation.gap_thresholds (
    kind   text NOT NULL PRIMARY KEY CHECK (kind IN ('prices', 'fx_rates', 'index_readings')),
    days   integer NOT NULL CHECK (days BETWEEN 1 AND 3660),
    set_at timestamptz NOT NULL DEFAULT now()
);
