---
title: Valuation totals per date are exact sums in PLN and per original currency
status: active
sourceFile: internal/views/valuation_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:17:58.833Z'
id: TEST-views-valuation
type: test
verification_scope: unit
---
Table tests run views.TotalsByDate over synthetic stored valuations: none gives no totals; three on one date in USD and PLN give 1451.20 PLN with 1050.00 PLN and 100.30 USD per currency, summed exactly as decimals; two dates in reverse order come out in date order. A PLN value in USD fails with ErrNotPLN, so currencies never mix. ValidateRange accepts one day and exactly 3700 days and rejects a reversed range and one day more with ErrInvalidRange. Results are compared in full with go-cmp; a float or a dropped currency would fail.