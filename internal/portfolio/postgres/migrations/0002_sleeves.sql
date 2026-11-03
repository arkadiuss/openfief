-- Sleeve kinds, sleeves and which account is in which sleeve belong to the portfolio context
-- (docs/dev/code-organization.md): the ledger knows nothing of sleeves. Accounts are named by the ledger's account
-- ID, as valuation.tax_wrappers names them; no foreign key crosses into the ledger schema. Nothing is copied from
-- the ledger schema's old sleeve tables: what was stored before the first release is entered again (D-161).

-- A named category of sleeve with a default role (author decision, sleeve kinds).
CREATE TABLE portfolio.sleeve_kinds (
    key          text PRIMARY KEY CHECK (key ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(key) <= 64),
    name         text NOT NULL CHECK (name <> ''),
    default_role text NOT NULL CHECK (default_role IN ('defensive', 'offensive'))
);

-- A kind can be renamed, but its default role never changes: sleeves were checked against it, and no change may
-- move capital into agents' reach (GR-6) without the owner overriding a sleeve explicitly.
CREATE FUNCTION portfolio.sleeve_kind_role_fixed() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.key <> OLD.key OR NEW.default_role <> OLD.default_role THEN
        RAISE EXCEPTION 'sleeve kind % keeps its key and default role', OLD.key
            USING ERRCODE = 'integrity_constraint_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER sleeve_kind_role_fixed BEFORE UPDATE ON portfolio.sleeve_kinds
    FOR EACH ROW EXECUTE FUNCTION portfolio.sleeve_kind_role_fixed();

-- The built-in kinds, as portfolio.DefaultKinds returns them.
INSERT INTO portfolio.sleeve_kinds (key, name, default_role) VALUES
    ('agent-pool', 'Agent pool', 'offensive'),
    ('bonds', 'Bonds', 'defensive'),
    ('core', 'Core', 'defensive'),
    ('emergency-fund', 'Emergency fund', 'defensive'),
    ('retirement', 'Retirement', 'defensive'),
    ('satellite', 'Satellite', 'offensive');

-- D-71: a sleeve whose role differs from its kind's default role needs the owner's explicit choice, either way:
-- role_override is set exactly then.
CREATE TABLE portfolio.sleeves (
    id            text PRIMARY KEY CHECK (id <> ''),
    name          text NOT NULL CHECK (name <> ''),
    kind          text NOT NULL REFERENCES portfolio.sleeve_kinds (key),
    role          text NOT NULL CHECK (role IN ('defensive', 'offensive')),
    role_override boolean NOT NULL DEFAULT false
);

-- The override is set exactly when the role differs from the kind's default role, so it always means the one thing
-- (portfolio.CheckSleeve). A kind's default role never changes, so checking the sleeve's row suffices.
CREATE FUNCTION portfolio.sleeve_role_override() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF EXISTS (SELECT FROM portfolio.sleeve_kinds k
               WHERE k.key = NEW.kind AND (NEW.role <> k.default_role) <> NEW.role_override) THEN
        RAISE EXCEPTION 'sleeve % has a role override exactly when its role differs from its kind''s', NEW.id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER sleeve_role_override BEFORE INSERT OR UPDATE ON portfolio.sleeves
    FOR EACH ROW EXECUTE FUNCTION portfolio.sleeve_role_override();

-- Which sleeve each account is in: one sleeve per account. A sleeve with accounts in it can't be deleted (D-120).
CREATE TABLE portfolio.account_sleeves (
    account_id text PRIMARY KEY CHECK (account_id <> ''),
    sleeve_id  text NOT NULL REFERENCES portfolio.sleeves (id)
);

CREATE INDEX account_sleeves_sleeve ON portfolio.account_sleeves (sleeve_id);
