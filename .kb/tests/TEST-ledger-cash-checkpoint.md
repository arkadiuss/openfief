---
title: A cash checkpoint's difference is a deposit or withdrawal on its date that makes it match
status: active
sourceFile: internal/ledger/cashcheckpoint_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:49:47.818Z'
id: TEST-ledger-cash-checkpoint
type: test
---
Table tests over the ledger's synthetic log, whose derived USD cash on 2026-03-06 is 1287: a checkpoint showing more is booked as a deposit and one showing less as a withdrawal, dated on the checkpoint and keyed by its date, currency and both balances; appending it makes the checkpoint match while the cash of the day before stays as derived; a matching checkpoint, an assertion a connector reported and a holding are refused. The attention item's subject and detail are pinned. A false pass would need CashDifference to book a wrong sign or a date other than the checkpoint's (D-158).