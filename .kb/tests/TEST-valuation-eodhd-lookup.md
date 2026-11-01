---
title: EODHD's exchange of a country and the one ISIN listed for a code are found
status: active
sourceFile: internal/valuation/eodhd/lookup_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:36.000Z'
id: TEST-valuation-eodhd-lookup
type: test
---
Pure table tests: CountryExchange maps each country with a supported exchange to it and every other to none, and FindByCode finds the one ISIN listed as a code on an exchange, ignoring case and repeats, and none when no listing or two ISINs match.