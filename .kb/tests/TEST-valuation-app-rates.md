---
title: Fetching FX rates splits a date range per provider limit and merges the results
status: active
sourceFile: internal/valuation/app/rates_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:54:11.843Z'
id: TEST-valuation-app-rates
type: test
---
Table tests cut synthetic date ranges with SplitRange: one day, exactly 93 days, 94 days (two chunks, the second one day long), a leap year in 150-day chunks and a range across a year end; a range that ends before it starts and a span limit of zero fail with ErrInvalidRange. FetchRates runs against an in-memory fake RateProvider holding synthetic rates: it asks for exactly the split ranges in date order, returns the rates sorted by date then currency, returns an empty result when the provider published nothing, and stops at the first provider failure, wrapping its error.