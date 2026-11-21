-- Whether the Ask box may give opinions (docs/DESIGN.md D-225, D-229): the owner's own setting, off by default. At most
-- one row; no row means it was never turned on and is off. Where it takes effect is decided in code: only with the one
-- switch and the Ask box's toggle on, for the owner's own session, never on a demo, and only when the route has a model
-- allowed for every data tier it uses.
CREATE TABLE modelrouter.opinions (
    singleton  boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    enabled    boolean NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
