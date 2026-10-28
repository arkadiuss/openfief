// Package postgres is the ledger's store in PostgreSQL (docs/DESIGN.md
// D-14): an app.Store over a pgx pool, and Migrate, which creates and
// upgrades the ledger schema from the SQL files embedded in migrations/.
//
// The ledger owns the ledger schema; no other context reads its tables. The
// transaction log is append-only in the database itself: triggers refuse any
// UPDATE, DELETE or TRUNCATE of transaction versions and postings, a version
// that is not the next one of its key, and postings added to a version
// stored earlier. Amounts and quantities are numeric, never floating point,
// and every version's fingerprint is checked when it is loaded.
//
// Store.InTx runs the ledger's writes in one database transaction, and
// Store.WithTx hands that transaction to work that must commit with them,
// such as inserting a River job (D-36).
//
// Migrate is a small hand-rolled runner instead of a library such as goose:
// it needs only pgx, keeps its version table (ledger.schema_migrations) and
// advisory lock key to the ledger, and applies pending migrations in one
// transaction. So it coexists with other contexts' migrations and with
// River's own (rivermigrate), which keep their own tables in their own
// schema.
package postgres
