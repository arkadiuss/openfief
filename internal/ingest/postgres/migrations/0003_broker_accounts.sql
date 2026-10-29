-- Broker accounts (docs/DESIGN.md D-72): the accounts each connector instance reported, as the institution names
-- them, and the OpenFief account the owner mapped each to. Unlike the rest of this schema they hold account numbers,
-- because the owner picks them here instead of typing them; they are shown to the owner and never logged (GR-12).
CREATE TABLE ingest.broker_accounts (
    instance         text NOT NULL CHECK (instance ~ '^[a-z]([-a-z0-9]*[a-z0-9])?$' AND length(instance) <= 56),
    broker_account   text NOT NULL CHECK (broker_account <> '' AND broker_account !~ '[[:space:][:cntrl:]]' AND length(broker_account) <= 64),
    institution      text NOT NULL CHECK (institution <> ''),
    -- NULL while unmapped: the instance's entries for it are not booked.
    account_id       text CHECK (account_id <> ''),
    mapped_at        timestamptz,
    -- The date of the last delivery that reported it, and whether the instance's last delivery that reported broker
    -- accounts reported it.
    last_seen        date NOT NULL,
    in_last_delivery boolean NOT NULL,
    -- An opening batch of the instance completed after it was mapped (D-23): the next one leaves its seeding out.
    seeded           boolean NOT NULL DEFAULT false,
    PRIMARY KEY (instance, broker_account),
    CHECK ((account_id IS NULL) = (mapped_at IS NULL)),
    CHECK (account_id IS NOT NULL OR NOT seeded)
);
