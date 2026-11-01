---
title: The ledger deletes only an account, sleeve or kind nothing refers to
status: active
sourceFile: internal/ledger/app/deletions_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:03.920Z'
id: TEST-ledger-app-deletions
type: test
---
Runs the ledger use cases over the in-memory store with the built-in kinds, the sleeve core and two accounts, one with a deposit: deleting that account, the sleeve core and its kind is refused with the sentinel errors; the empty account and an unused built-in kind are deleted, deleting again changes nothing, and HasTransactionsFrom matches by key source prefix.