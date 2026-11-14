-- The model-portfolio maintainer's runs, notes and proposals (docs/DESIGN.md D-204). Only the portfolio context reads or
-- writes them. A run is one ask of the maintainer's model: queued when it is asked for, running while its job works,
-- then final. It stores the model's note and nothing else the model or the owner's figures made: no prompt, no answer.

-- One row per run. At most one run is queued or running at a time (the unique index): the maintainer works one run at
-- a time for the one model portfolio. updated_at is when the run was last queued, started or ended, which is also when
-- an abandoned run is told apart from a live one.
CREATE TABLE portfolio.maintainer_runs (
    id          text PRIMARY KEY CHECK (id ~ '^[0-9a-f]{16}$'),
    trigger     text NOT NULL CHECK (trigger IN ('weekly', 'owner')),
    state       text NOT NULL CHECK (state IN ('queued', 'running', 'ok', 'refused', 'failed')),
    -- The owner's session that asked, empty for the schedule.
    actor       text CHECK (actor IS NULL OR actor ~ '^[0-9a-f]{16}$'),
    updated_at  timestamptz NOT NULL,
    -- The class a refused or failed run ended in, from a closed set of the audit log's.
    error_class text CHECK (error_class IS NULL OR error_class IN (
        'unavailable', 'rejected', 'canceled', 'misconfigured', 'internal', 'denied', 'invalid_output', 'no_figures')),
    -- The model that wrote the note, as "provider/model", and the note: plain text of 1 to 2000 characters.
    model       text CHECK (model IS NULL OR (char_length(model) BETWEEN 3 AND 192)),
    note        text CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 2000),
    CHECK ((state IN ('refused', 'failed')) = (error_class IS NOT NULL)),
    CHECK ((state = 'ok') = (note IS NOT NULL AND model IS NOT NULL))
);

CREATE UNIQUE INDEX maintainer_runs_one_active ON portfolio.maintainer_runs ((true)) WHERE state IN ('queued', 'running');
CREATE INDEX maintainer_runs_latest ON portfolio.maintainer_runs (updated_at DESC, id DESC);

-- The set of target changes a run proposed, which passed every rule (portfolio.CheckChanges). It is named by its run's
-- ID. pending waits in the Inbox; applied and dismissed are the owner's decisions; superseded is a pending one a later
-- run's proposal replaced.
CREATE TABLE portfolio.maintainer_proposals (
    run_id     text PRIMARY KEY REFERENCES portfolio.maintainer_runs (id),
    status     text NOT NULL CHECK (status IN ('pending', 'applied', 'dismissed', 'superseded')),
    created_at timestamptz NOT NULL,
    decided_at timestamptz,
    CHECK ((status IN ('applied', 'dismissed')) = (decided_at IS NOT NULL))
);

-- Its changes in the model's order, with the figures behind each: the weights in percent with at most two decimal
-- places, and an index, which can't be held, without a current weight or a drift.
CREATE TABLE portfolio.maintainer_changes (
    run_id         text NOT NULL REFERENCES portfolio.maintainer_proposals (run_id),
    position       integer NOT NULL CHECK (position >= 0),
    target         text NOT NULL CHECK (char_length(target) BETWEEN 3 AND 80),
    before_weight  numeric NOT NULL CHECK (before_weight > 0 AND before_weight <= 100 AND scale(before_weight) <= 2),
    after_weight   numeric NOT NULL CHECK (after_weight > 0 AND after_weight <= 100 AND scale(after_weight) <= 2),
    held           boolean NOT NULL,
    current_weight numeric CHECK (current_weight IS NULL OR scale(current_weight) <= 2),
    drift          numeric CHECK (drift IS NULL OR scale(drift) <= 2),
    reason         text NOT NULL CHECK (char_length(reason) BETWEEN 1 AND 200 AND reason !~ E'[\\r\\n]'),
    PRIMARY KEY (run_id, position),
    CHECK (held = (current_weight IS NOT NULL AND drift IS NOT NULL))
);
