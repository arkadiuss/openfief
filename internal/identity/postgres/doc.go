// Package postgres is the identity context's store in PostgreSQL
// (docs/DESIGN.md D-14): an app.Store for script tokens (D-27) over a pgx
// pool, and Migrate, which creates and upgrades the identity schema from the
// SQL files embedded in migrations/.
//
// The identity context owns the identity schema; no other context reads its
// tables. A script token row holds the SHA-256 hash of the token's key, never
// the key or the secret (D-29), so a database dump can't be replayed against
// the API. Migrate is the ledger's hand-rolled runner, copied with its own
// version table (identity.schema_migrations) and advisory lock key.
package postgres
