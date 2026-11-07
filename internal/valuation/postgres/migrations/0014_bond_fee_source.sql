-- Where each series' early-redemption fee came from (docs/DESIGN.md D-97): "entered" by the owner, or confirmed by
-- the owner when it was assumed; "table", the reviewed fee table's own fee for the series' sale month; or "assumed",
-- the table's latest fee for the bond type carried forward to a series sold after the month the table was reviewed
-- through, which raises an attention item until the owner confirms or overrides it. Terms stored before this
-- migration whose attribution says the fee came from the reviewed table get "table"; every other row was entered.
ALTER TABLE valuation.bond_terms
    ADD COLUMN fee_source text NOT NULL DEFAULT 'entered'
        CONSTRAINT bond_terms_fee_source_check CHECK (fee_source IN ('entered', 'table', 'assumed'));

UPDATE valuation.bond_terms SET fee_source = 'table'
WHERE source_processing LIKE '%OpenFief took it from its reviewed table of the issue letters';
