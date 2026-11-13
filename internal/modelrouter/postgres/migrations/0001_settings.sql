-- The modelrouter context's schema: the owner's agents-and-models settings (docs/DESIGN.md D-212, D-205, D-206). Only the
-- modelrouter context reads or writes it.

-- The one switch for agents and models. At most one row; no row means it was never saved and is off (D-212: off by
-- default).
CREATE TABLE modelrouter.switch (
    singleton  boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    enabled    boolean NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- The features whose toggle is on, under the switch (D-212). A feature with no row is off.
CREATE TABLE modelrouter.feature_toggles (
    feature    text PRIMARY KEY CHECK (feature IN ('ask', 'maintainer', 'mcp')),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Each role's models in the order they are tried: the primary at position 0, then the fallbacks (D-205). A
-- reference is a model-provider plugin instance's name and a model entry's id at that provider.
CREATE TABLE modelrouter.routes (
    role     text    NOT NULL CHECK (role IN ('ask', 'maintainer')),
    position integer NOT NULL CHECK (position >= 0),
    provider text    NOT NULL CHECK (provider <> ''),
    model    text    NOT NULL CHECK (model <> ''),
    PRIMARY KEY (role, position),
    UNIQUE (role, provider, model)
);

-- The models the owner allows to receive each data tier (D-206). A tier or a model with no row is denied.
CREATE TABLE modelrouter.tier_models (
    tier     text    NOT NULL CHECK (tier IN ('T0', 'TF', 'T1', 'T2')),
    position integer NOT NULL CHECK (position >= 0),
    provider text    NOT NULL CHECK (provider <> ''),
    model    text    NOT NULL CHECK (model <> ''),
    PRIMARY KEY (tier, position),
    UNIQUE (tier, provider, model)
);
