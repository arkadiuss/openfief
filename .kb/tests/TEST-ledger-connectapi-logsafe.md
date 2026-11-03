---
title: The ledger services' log sinks write no values
status: active
verification_scope: unit
sourceFile: internal/ledger/connectapi/logsafe_internal_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T03:24:10.700Z'
id: TEST-ledger-connectapi-logsafe
type: test
---
Drives every log sink of the ledger context's Connect services with failures whose text carries a synthetic amount, ISIN and IBKR-style account number, a store error quoting a row, joined errors and a deadline among them, into a text logger, and checks every line against a deny-list regex of amounts, ISINs and account numbers. A sink logging err.Error() fails it.