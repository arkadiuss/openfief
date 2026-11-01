---
title: A confirmed transfer link breaks when a correction makes its legs not match
status: active
sourceFile: internal/ledger/transferbroken_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:53.758Z'
id: TEST-ledger-transfer-broken
type: test
---
Pure table test of ledger.BrokenTransferLinks over synthetic legs: a matching pair is not broken; a corrected amount, a date moved out of the five-day window and a deleted leg break a confirmed link; a rejected pair is never broken.