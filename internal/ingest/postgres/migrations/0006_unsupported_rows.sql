-- How many rows of a completed batch OpenFief could not book yet (ingest.ProblemUnsupported), apart from the rows
-- left out on purpose that problems also counts: each batch with any raises an attention item (docs/DESIGN.md D-130).
-- NULL for batches completed before this column existed, and for open ones.
ALTER TABLE ingest.import_batches ADD COLUMN unsupported integer CHECK (unsupported >= 0 AND unsupported <= problems);
