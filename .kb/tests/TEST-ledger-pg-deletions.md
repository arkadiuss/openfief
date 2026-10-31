---
title: In PostgreSQL, a delete checks and deletes in one statement and the transaction log keeps every version
status: active
sourceFile: internal/ledger/postgres/deletions_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:05.426Z'
id: TEST-ledger-pg-deletions
type: test
---
Runs the ledger's Postgres store on a fresh database: the account holding a deposit, its sleeve and the kind are refused; the empty account and an unused kind go; the deposit's version is still stored; HasTransactionsFrom matches by source prefix; and once both accounts are gone the sleeve and then its kind are deleted.