// Package postgres is the portfolio context's store in PostgreSQL
// (docs/DESIGN.md D-14): the optional model portfolio (D-52), whether it is
// on and its targets, over a pgx pool, and Migrate, which creates and
// upgrades the portfolio schema from the SQL files embedded in migrations/.
//
// The portfolio context owns the portfolio schema; no other context reads
// its tables. Migrate is the ledger's hand-rolled runner, copied with its own
// version table (portfolio.schema_migrations) and advisory lock key.
package postgres
