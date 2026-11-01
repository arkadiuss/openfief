---
title: FX rates are looked up on or before a date and convert to PLN with explicit rounding
status: active
sourceFile: internal/valuation/fx_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:45.433Z'
id: TEST-valuation-fx
type: test
verification_scope: unit
---
Table tests use a synthetic NBP-style table (USD on Friday 2026-03-06 and Monday 2026-03-09, EUR on Monday): a same-day lookup takes Monday, a Sunday or Saturday cutoff takes Friday, a later table is never used, and a cutoff before the first table or an absent currency fails with ErrNoRate; two different rates for one date fail with ErrConflictingRates. With the Polish tax cutoff, income on Monday converts at Friday's rate. Conversion multiplies exactly and rounds once to grosze with the caller's mode (495.318435 gives 495.32 half up and 495.31 down; the 45.225 tie splits half up and half even), rejects a rate for another currency with a CurrencyMismatchError and a zero rate, and passes PLN through with rounding only.