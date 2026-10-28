-- Sleeve kinds: a named category of sleeve with a default role (author
-- decision, sleeve kinds). Like sleeves, they live in the ledger schema
-- because the portfolio context has no store of its own in M0.

CREATE TABLE ledger.sleeve_kinds (
    key          text PRIMARY KEY CHECK (key ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(key) <= 64),
    name         text NOT NULL CHECK (name <> ''),
    default_role text NOT NULL CHECK (default_role IN ('defensive', 'offensive'))
);

-- A kind can be renamed, but its default role never changes: sleeves were
-- checked against it, and no change may move capital into agents' reach
-- (GR-6) without the owner overriding a sleeve explicitly.
CREATE FUNCTION ledger.sleeve_kind_role_fixed() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.key <> OLD.key OR NEW.default_role <> OLD.default_role THEN
        RAISE EXCEPTION 'sleeve kind % keeps its key and default role', OLD.key
            USING ERRCODE = 'integrity_constraint_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER sleeve_kind_role_fixed BEFORE UPDATE ON ledger.sleeve_kinds
    FOR EACH ROW EXECUTE FUNCTION ledger.sleeve_kind_role_fixed();

-- The built-in kinds, as portfolio.DefaultKinds returns them.
INSERT INTO ledger.sleeve_kinds (key, name, default_role) VALUES
    ('agent-pool', 'Agent pool', 'offensive'),
    ('bonds', 'Bonds', 'defensive'),
    ('core', 'Core', 'defensive'),
    ('emergency-fund', 'Emergency fund', 'defensive'),
    ('retirement', 'Retirement', 'defensive'),
    ('satellite', 'Satellite', 'offensive');

-- Every sleeve has a kind. Sleeves stored before kinds existed get core when
-- defensive and satellite when offensive: the kinds whose default role is
-- theirs, so none needs an override.
ALTER TABLE ledger.sleeves
    ADD COLUMN kind text REFERENCES ledger.sleeve_kinds (key),
    ADD COLUMN offensive_override boolean NOT NULL DEFAULT false
        CHECK (NOT offensive_override OR role = 'offensive');

UPDATE ledger.sleeves SET kind = CASE role WHEN 'offensive' THEN 'satellite' ELSE 'core' END;

ALTER TABLE ledger.sleeves ALTER COLUMN kind SET NOT NULL;
