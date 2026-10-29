---
title: Valuation bases and market valuations
status: active
sourceFile: internal/valuation/valuation_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:48.205Z'
id: TEST-valuation-valuation
type: test
---
Basis names parse back (market, gross_accrued; net_if_redeemed is no basis since D-77) and a market valuation is quantity times close, exact and in the price's currency; a price of another instrument is refused.