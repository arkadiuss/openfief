---
title: Holding valuations with their three values, missing inputs, valued-through dates and coalescing
status: active
sourceFile: internal/valuation/revaluation_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:20.208Z'
id: TEST-valuation-revaluation
type: test
---
Pure tests with synthetic prices, rates and lots: a USD ETF worth 5084.09 PLN on a cost of 4800.00 PLN has its net and liquidation values in each wrapper and under fee rules (commission, conversion fee, a USD minimum converted on the day), and lacks them with the reason when a minimum or a lot's cost has no rate, its lots are unknown or the owner's marginal rate is not set; bond holdings sum their lots' three values per wrapper; missing inputs become rows; valued-through dates and coalescing.