---
title: An import hook inserts jobs on the import's Postgres transaction
status: active
sourceFile: internal/ledger/postgres/hook_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:28:56.233Z'
id: TEST-ledger-pg-hook
type: test
---
On a fresh database with a stand-in public.jobs table, an import hook gets the transaction's *postgres.Store and inserts one row per entry of EarliestAffected through WithTx. When the hook succeeds the job row acc-1 2026-03-02 and the deposit are both committed; when the hook fails after inserting, Import returns the error and neither the job nor the deposit is stored. WithTx on a Store outside InTx fails with ErrNoTx.