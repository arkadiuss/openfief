---
title: The views list suggested, confirmed and rejected transfers across all accounts
status: active
sourceFile: internal/views/app/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:01.506Z'
id: TEST-views-app-transfers
type: test
verification_scope: unit
---
Over fake ledger ports holding two withdrawals in acc-1 and two deposits of the same amounts in acc-2, the Transfers query suggests both pairs while nothing is decided, and with one pair confirmed and one rejected it suggests nothing and lists each pair under its decision; a failing ledger fails the query. A false pass would need the fake to hide transactions.