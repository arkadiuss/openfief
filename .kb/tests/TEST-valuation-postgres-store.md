---
title: Inputs and holding valuations round-trip through the valuation schema
status: active
sourceFile: internal/valuation/postgres/store_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:55:58.194Z'
id: TEST-valuation-postgres-store
type: test
---
On a fresh database, synthetic prices, an FX rate, a CPI reading, EDO series terms and an IKE tax wrapper are saved and read back exactly; upserting a price replaces it; a price after the cutoff is not returned; an ISIN without terms is not a bond; an account without a wrapper is taxable. ReplaceValuations stores valued and missing-input rows, replaces only those from its date on, and records the valued-through date; Valuations reads them back in date order.