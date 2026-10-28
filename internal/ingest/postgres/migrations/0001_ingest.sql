-- Connector syncs (docs/DESIGN.md D-49): each connector instance's sync state and watermark, the history of its
-- import batches, and its connector bindings. Rows are kept when an instance is uninstalled (D-58), so reinstalling
-- it under the same name continues from its watermark. Nothing here holds an amount, a holding or an account number:
-- only internal IDs, counts and times (GR-12).
CREATE TABLE ingest.import_batches (
    id                text PRIMARY KEY CHECK (id <> ''),
    instance          text NOT NULL CHECK (instance ~ '^[a-z]([-a-z0-9]*[a-z0-9])?$' AND length(instance) <= 56),
    source            text NOT NULL CHECK (source IN ('sync', 'file')),
    mode              text NOT NULL CHECK (mode IN ('opening', 'transactions')),
    started_at        timestamptz NOT NULL,
    -- open until it completes; abandoned when the owner started another batch in its place.
    state             text NOT NULL DEFAULT 'open' CHECK (state IN ('open', 'completed', 'abandoned')),
    attempts          integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_attempt_at   timestamptz,
    finished_at       timestamptz,
    -- What the completed batch imported (ingest.Batch); NULL until it completes.
    institution       text,
    transactions      integer CHECK (transactions >= 0),
    assertions        integer CHECK (assertions >= 0),
    new               integer CHECK (new >= 0),
    new_versions      integer CHECK (new_versions >= 0),
    unchanged         integer CHECK (unchanged >= 0),
    failed_assertions integer CHECK (failed_assertions >= 0),
    problems          integer CHECK (problems >= 0),
    CHECK ((state = 'completed') = (finished_at IS NOT NULL AND institution IS NOT NULL))
);

CREATE INDEX import_batches_instance ON ingest.import_batches (instance, started_at);

CREATE TABLE ingest.connector_instances (
    instance     text PRIMARY KEY CHECK (instance ~ '^[a-z]([-a-z0-9]*[a-z0-9])?$' AND length(instance) <= 56),
    -- What the last successful sync returned; opaque to the core.
    watermark    text NOT NULL DEFAULT '',
    seeded       boolean NOT NULL DEFAULT false,
    failures     integer NOT NULL DEFAULT 0 CHECK (failures >= 0),
    -- NULL until the first attempt ends: due at once.
    next_sync_at timestamptz,
    -- The claim of the attempt that runs now; expired or NULL when none runs.
    lease_until  timestamptz,
    open_batch   text REFERENCES ingest.import_batches (id)
);

CREATE TABLE ingest.connector_bindings (
    instance   text NOT NULL CHECK (instance ~ '^[a-z]([-a-z0-9]*[a-z0-9])?$' AND length(instance) <= 56),
    account_id text NOT NULL CHECK (account_id <> ''),
    bound_at   timestamptz NOT NULL,
    PRIMARY KEY (instance, account_id)
);
