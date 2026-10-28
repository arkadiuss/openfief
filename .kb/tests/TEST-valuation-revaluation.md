---
title: Holding valuations, missing inputs, valued-through dates and coalescing
status: active
sourceFile: internal/valuation/revaluation_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:20.208Z'
id: TEST-valuation-revaluation
type: test
---
Table tests value synthetic holdings: 12.5 ETF units on a Sunday at Friday's 101.37 USD close and 4.0123 rate give exactly 1267.125 USD and 5084.09 PLN half up; a PLN close needs no rate; no price or no rate yields a row marked missing with no value instead of an error, and conflicting rates fail. A two-lot EDO holding sums pl.ValueLot's net values by default and gross values when configured, takes its second-year CPI reading, and is marked missing without it; the market basis, a non-PLN cost and a lot outside its series' month fail. ValuedThrough stops the day before the first gap, keeps an earlier gap before the start, and Coalesce keeps the earliest date per account.