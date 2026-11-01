---
title: An account's payout setting pays out only from the day it took effect
status: active
sourceFile: internal/ledger/payout_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:11.835Z'
id: TEST-ledger-payout
type: test
---
Table tests of ledger.PayoutSetting: the default keeps payouts as cash; a paid-out setting pays out what falls due on or after its day and not before; changing to the same setting keeps its day, another one takes today; keys parse and print as cash and paid-out, and an unknown key or value is refused. A false pass would need PaidOut to ignore the day, which the before-the-change case catches.