// Package postgres is the ingest context's store in PostgreSQL
// (docs/DESIGN.md D-14): the app.SyncStore that keeps each connector
// instance's watermark, sync state and import batch history (D-49), and the
// app.BindingStore of connector bindings, over a pgx pool. Migrate creates
// and upgrades the ingest schema from the SQL files embedded in
// migrations/.
//
// The ingest context owns the ingest schema; no other context reads its
// tables. Rows hold internal IDs, counts and times only, never an amount or
// an account number (GR-12). A claim on an instance is a row-level lease
// with an expiry, so two core replicas never sync one instance at once and a
// crashed attempt's claim lapses. Migrate is the ledger's hand-rolled runner,
// copied with its own version table (ingest.schema_migrations) and advisory
// lock key.
package postgres
