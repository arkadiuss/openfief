// Package postgres is the thesis context's store in PostgreSQL
// (docs/DESIGN.md D-14): the owner's draft theses and every version of each
// (D-229), over a pgx pool, and Migrate, which creates and upgrades the thesis
// schema from the SQL files embedded in migrations/.
//
// The thesis context owns the thesis schema; no other context reads its
// tables. Migrate is the ledger's hand-rolled runner, copied with its own
// version table (thesis.schema_migrations) and advisory lock key.
package postgres
