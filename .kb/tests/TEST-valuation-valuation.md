---
title: Valuation bases, market valuations and the bond basis for totals
status: active
sourceFile: internal/valuation/valuation_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:48.205Z'
id: TEST-valuation-valuation
type: test
---
Table tests parse and print the three basis names and reject other spellings; value 12.5 units at a synthetic end-of-day close of 101.37 USD as exactly 1267.125 USD on the market basis and reject a price of another instrument; resolve the bond basis (unset gives net_if_redeemed, gross_accrued and net_if_redeemed are kept, market and unknown values fail); and pick the value of a synthetic EDO lot computed by pl.ValueLot: 1008.26 PLN net by default, 1030.20 PLN gross when configured.