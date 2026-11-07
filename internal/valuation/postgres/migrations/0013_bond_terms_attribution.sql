-- Series terms the treasury-bonds connector read from the Ministry of Finance's published retail bond data carry the
-- attribution its licence (CC BY 4.0) asks for: the source, the dataset and its licence, the file, when the ministry
-- made the data and when it was taken, and how it was processed (docs/DESIGN.md D-97). Terms the owner entered carry
-- none, so either every part is there or none is; the latest entry replaces the row, its attribution included.
ALTER TABLE valuation.bond_terms
    ADD COLUMN source_publisher  text,
    ADD COLUMN source_dataset    text,
    ADD COLUMN source_license    text,
    ADD COLUMN source_url        text,
    ADD COLUMN source_made       date,
    ADD COLUMN source_fetched    timestamptz,
    ADD COLUMN source_processing text,
    ADD CONSTRAINT bond_terms_attribution_check CHECK (
        (source_publisher IS NULL AND source_dataset IS NULL AND source_license IS NULL AND source_url IS NULL
            AND source_made IS NULL AND source_fetched IS NULL AND source_processing IS NULL)
        OR (source_publisher <> '' AND source_dataset <> '' AND source_license <> '' AND source_url <> ''
            AND source_made IS NOT NULL AND source_fetched IS NOT NULL AND source_processing <> '')
    );
