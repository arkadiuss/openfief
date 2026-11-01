---
title: Market-data series, their one source, carry-forward and gaps, in the domain
status: active
sourceFile: internal/valuation/series_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:05.587Z'
id: TEST-valuation-series
type: test
---
Table tests of internal/valuation/series.go with synthetic prices, rates and readings: a series' key and kind validate (an instrument, a currency other than PLN, an index kind); ChooseSource picks the highest-ranked source that has a value, passing over a higher one without values and ranking unnamed sources after named ones by name; PricesBySource, FXRatesBySource and IndexReadingsBySource keep only the series source's values and put every other source's aside; a day only another source priced takes the series source's latest earlier price through PriceOnOrBefore, never the other price; OtherSourceChecks turns the others into cross-check prices; gap thresholds default to 7 and 92 days and validate 1 to 3660; GapOf flags a gap only beyond the threshold, a weekend or exactly the threshold being none.