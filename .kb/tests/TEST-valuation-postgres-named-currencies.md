---
title: Stored valuations and rates round-trip the currency they name
status: active
verification_scope: integration
sourceFile: internal/valuation/postgres/currency_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T23:54:38.292Z'
id: TEST-valuation-postgres-named-currencies
type: test
---
Against a fresh PostgreSQL database with every migration applied, stores a valuation in EUR, as a pack whose reporting currency is EUR would, and FX rates of USD into PLN and into EUR, and reads them back: the valuation names EUR for its gross, net, cost basis and sale cost, and each rate its own base. It also checks that no constraint of the valuation schema mentions PLN, and that a rate quoted in its own currency is refused. A codec that assumed PLN fails it.