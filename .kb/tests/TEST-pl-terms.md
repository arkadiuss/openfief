---
title: Interest period rates come from the first-period rate, the index and the margin
status: active
sourceFile: internal/jurisdiction/pl/terms_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:13.668Z'
id: TEST-pl-terms
type: test
---
Table tests take synthetic series terms and index readings and check the yearly rate of a period: period 1 uses the fixed first-period rate, EDO period 2 is CPI 4.9% plus margin 2% = 6.9%, a CPI of -1% counts as zero so the rate is the margin, ROR and DOR use the NBP reference rate plus margin, and TOS stays fixed. A missing reading returns a MissingIndexError naming the series, period and start date. Rates are parsed from text and compared with Decimal.Equal.