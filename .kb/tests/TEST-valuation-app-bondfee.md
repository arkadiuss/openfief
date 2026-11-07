---
title: The owner confirms an assumed bond fee or overrides any fee through the settings use cases
status: active
sourceFile: internal/valuation/app/bondfee_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:22.325Z'
id: TEST-valuation-app-bondfee
type: test
---
TestSetEarlyRedemptionFee drives Settings.SetEarlyRedemptionFee over a fake BondFees store holding one series' terms: confirming an assumed fee saves it as entered with no revaluation, overriding a table's fee saves the new fee and revalues the account that held the series from its first transaction, the same amount revalues nothing, and confirming a fee that isn't assumed, a bond without terms, a negative or non-numeric fee, a malformed ISIN and a read session are refused with nothing saved.