---
title: Revalue rebuilds an account's holding valuations, three values each, from its log and stored inputs
status: active
sourceFile: internal/valuation/app/revalue_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:23.924Z'
id: TEST-valuation-app-revalue
type: test
verification_scope: unit
---
Use-case tests over in-memory fakes: rows from the log and stored inputs with gross, net and liquidation values; idempotent; starts at the first transaction; keeps an earlier gap; bonds valued in the recorded wrapper, the pack's default when none is recorded, with the owner's marginal rate from the store; the account's fee rule reaches liquidation; a short holding is stored with its net and liquidation values missing lots.