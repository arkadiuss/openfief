// Package postgres is the modelrouter context's store in PostgreSQL
// (docs/DESIGN.md D-14): the owner's agents-and-models settings (D-212,
// D-205, D-206), the one switch, the feature toggles, each role's route and
// the models allowed per data tier, over a pgx pool, and Migrate, which
// creates and upgrades the modelrouter schema from the SQL files embedded in
// migrations/.
//
// The modelrouter context owns the modelrouter schema; no other context reads its
// tables. Migrate is the ledger's hand-rolled runner, copied with its own
// version table (modelrouter.schema_migrations) and advisory lock key.
package postgres
