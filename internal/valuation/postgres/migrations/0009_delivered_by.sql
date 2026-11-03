-- Bond series terms and market index levels come only from a connector whose signed manifest declares them and whose
-- grant approves that, and every stored term and index level records the connector instance that delivered it
-- (docs/DESIGN.md D-179). Index levels already keep their source; it may no longer be empty.
--
-- Terms stored before this migration name no instance, and nothing stored before the first release is converted
-- (D-161): they go, and the treasury-bonds connector delivers them again with the next lot the owner enters. The
-- demo reseeds itself.
DELETE FROM valuation.bond_terms;
ALTER TABLE valuation.bond_terms
    ADD COLUMN source text NOT NULL CHECK (source <> '' AND length(source) <= 64);

DELETE FROM valuation.index_levels WHERE source = '';
ALTER TABLE valuation.index_levels
    ADD CONSTRAINT index_levels_source_check CHECK (source <> '');
