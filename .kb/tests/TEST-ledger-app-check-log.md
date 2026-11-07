---
title: The ledger use case checks every stored account log together
status: active
sourceFile: internal/ledger/app/checklog_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:26.678Z'
id: TEST-ledger-app-check-log
type: test
---
TestCheckLog runs Ledger.CheckLog over the in-memory store holding a synthetic history in two accounts, a deposit corrected from one account to the other included, and checks the counts per account, the moved version counted once as superseded in the first account, and a balanced log summing to zero in each currency.