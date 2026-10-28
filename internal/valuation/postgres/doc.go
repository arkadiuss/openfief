// Package postgres is the valuation context's store and job queue in
// PostgreSQL (docs/DESIGN.md D-14, D-36): prices, FX rates, period index
// readings, bond series terms and tax wrappers; the stored valuations and
// each account's valued-through date; and the River queue that revalues
// accounts.
//
// The valuation context owns the valuation schema, and River's tables live
// in the river schema; Migrate creates and upgrades both. Accounts and
// instruments refer to the ledger by value only.
//
// Every write of an input enqueues the revaluations it causes in the same
// database transaction, and so does every ledger import, through
// Store.ImportHook on the import's own transaction (D-36). Pending
// revaluations coalesce per account: River keeps at most one RevalueAccount
// job per account waiting or running (unique by args), and valuation.dirty
// holds the earliest affected date, lowered with LEAST() by every enqueue.
// RevalueWorker claims that date, revalues, and clears it in the same
// transaction that completes the job, under a per-account advisory lock that
// enqueueing takes too, so a change that arrives during a run is never left
// without a job.
package postgres
