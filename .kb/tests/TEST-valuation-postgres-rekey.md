---
title: Valuation's stored inputs of a bond series move to its registry ISIN, the registry's own winning
status: active
sourceFile: internal/valuation/postgres/rekey_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:19.440Z'
id: TEST-valuation-postgres-rekey
type: test
---
Against a disposable PostgreSQL (task test:pg), stores a series' terms, exemption class, prices, cross-check price, threshold and prices source under its provisional ISIN, a price under the registry ISIN and another series' terms; the re-key moves them all, keeps the registry's own price for a day both have and the other series, changes nothing run again, and drops terms stored again under the provisional ISIN once the registry's exist.