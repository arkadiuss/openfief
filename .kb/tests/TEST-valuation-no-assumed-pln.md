---
title: No valuation or view code, query or current schema assumes PLN
status: active
verification_scope: unit
sourceFile: internal/valuation/nopln_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T23:54:36.861Z'
id: TEST-valuation-no-assumed-pln
type: test
---
Reads every non-test Go file of the valuation, valuation/app, valuation/postgres, views, views/app and portfolio/app packages and fails on the literal PLN, outside the NBP adapter and the Polish pack, which quote against PLN by definition. A conversion, a codec or a query that names PLN, or a currency constant spelled out again, fails it.