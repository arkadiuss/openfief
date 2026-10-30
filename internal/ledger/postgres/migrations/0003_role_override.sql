-- D-71: a sleeve whose role differs from its kind's default role needs the owner's explicit choice, either way. The
-- offensive override of 0002 becomes a role override. Sleeves stored before were valid when they were defensive in an
-- offensive kind with no override; they keep their role and get the override, as the explicit choice they were then.
ALTER TABLE ledger.sleeves RENAME COLUMN offensive_override TO role_override;
-- 0002's CHECK (NOT offensive_override OR role = 'offensive'), named by PostgreSQL.
ALTER TABLE ledger.sleeves DROP CONSTRAINT sleeves_check;

UPDATE ledger.sleeves s SET role_override = (s.role <> k.default_role)
    FROM ledger.sleeve_kinds k WHERE k.key = s.kind;

-- The override is set exactly when the role differs from the kind's default role, so it always means the one thing
-- (portfolio.CheckSleeve). A kind's default role never changes (0002), so checking the sleeve's row suffices.
CREATE FUNCTION ledger.sleeve_role_override() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF EXISTS (SELECT FROM ledger.sleeve_kinds k
               WHERE k.key = NEW.kind AND (NEW.role <> k.default_role) <> NEW.role_override) THEN
        RAISE EXCEPTION 'sleeve % has a role override exactly when its role differs from its kind''s', NEW.id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER sleeve_role_override BEFORE INSERT OR UPDATE ON ledger.sleeves
    FOR EACH ROW EXECUTE FUNCTION ledger.sleeve_role_override();
