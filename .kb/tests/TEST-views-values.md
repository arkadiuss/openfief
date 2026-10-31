---
title: Holdings and cash give each of the three values per kind, and totals sum them with cash
status: active
sourceFile: internal/views/values_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:41.990Z'
id: TEST-views-values
type: test
verification_scope: unit
---
Pure tests with a synthetic IKZE row of 1000 PLN gross, 900 net and no liquidation value without the marginal rate: HoldingValue and CashValue per kind and their missing reasons; TotalsByDate with IKZE cash, cash without an FX rate and a later day, leaving the liquidation total unset; CashOnDays over a deposit, a PLN deposit and a buy; ValueHoldings of the net and liquidation kinds.