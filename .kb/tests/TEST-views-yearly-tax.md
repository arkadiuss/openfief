---
title: An OKI account's days for the yearly tax, per class at the day-before rate
status: active
sourceFile: internal/views/yearlytax_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:07.954Z'
id: TEST-views-yearly-tax
type: test
---
views.TaxDays on a synthetic OKI account with a treasury bond, a USD fund nobody classified, a PLN fund the owner classified, PLN and USD cash and payments checks each day's values per class against hand-computed amounts converted at the rate of the day before, not the day's, that opening cash is no payment, the unclassified instrument, and the error without rates.