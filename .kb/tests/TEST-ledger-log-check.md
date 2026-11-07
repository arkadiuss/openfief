---
title: The whole-log check sums every version per currency, names unbalanced ones and counts per account
status: active
sourceFile: internal/ledger/logcheck_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:25.858Z'
id: TEST-ledger-log-check
type: test
---
TestCheckLog builds stored logs as data, balanced or not, and checks ledger.CheckLog: counts of transactions, versions and superseded versions per account and in all, a deletion as a version, a version listed twice counted once, every unbalanced version named per currency with its sum, the per-currency sums of the whole log, every version listed with its key, kind and superseded mark, and Balanced only when nothing is unbalanced.