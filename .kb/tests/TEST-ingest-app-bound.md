---
title: Connectors refuse deliveries for unbound accounts and seed only in opening batches
status: active
sourceFile: internal/ingest/app/bound_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:49:58.402Z'
id: TEST-ingest-app-bound
type: test
---
Runs Connectors.Sync over fake callers, ledgers, stores and bindings. A delivery in a bound account is imported; with no binding, a binding to another account, or a failing binding store, nothing reaches the ledger or the valuation stores and the error wraps ErrAccountNotBound or the store's error; a delivery without transactions or assertions needs no binding lookup. An opening batch goes to the ledger's Seed, a transactions batch to Import, and a transactions batch with opening cash is refused with ErrSeedingOutsideOpening before any ledger call.