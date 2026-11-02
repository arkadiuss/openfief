---
title: The ledger withdraws holding assertions a ticker correction replaces
status: active
sourceFile: internal/ledger/app/withdraw_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T06:59:55.804Z'
id: TEST-ledger-app-withdraw
type: test
---
Over the in-memory store: a holding assertion and a cash assertion are checked; withdrawing the cash one is refused and removes nothing; withdrawing the holding removes its check, twice with no effect the second time, and leaves the cash assertion. A false pass would need a withdrawal to touch the owner's cash or the log (D-153).