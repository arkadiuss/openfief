-- The index readings bond terms depend on, such as CPI and the NBP reference rate, come only from a connector whose
-- signed manifest declares them and whose grant approves that, and each stored reading records the connector
-- instance that delivered it (docs/DESIGN.md D-186), as bond terms and index levels do (D-179). Readings already keep
-- their source; it may no longer be empty, and nothing stored before the first release is converted (D-161).
DELETE FROM valuation.index_readings WHERE source = '';
ALTER TABLE valuation.index_readings
    ADD CONSTRAINT index_readings_source_check CHECK (source <> '' AND length(source) <= 64);
