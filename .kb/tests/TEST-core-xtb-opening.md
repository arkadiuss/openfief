---
title: An XTB opening upload starts the account and confirmed positions join it in their trading currency
status: active
sourceFile: cmd/core/xtbopening_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T16:47:56.832Z'
id: TEST-core-xtb-opening
type: test
---
Runs the real core against a disposable PostgreSQL with a passkey owner session, the demo registry and an in-process connector that maps synthetic exports with xtb.Mapper and proposes TESTA.PL on a PLN listing. The opening upload books nothing with both positions waiting, the next upload is a transactions batch, and after the owner confirms both tickers the read API shows two opening lots on the report's date costing 864.15 PLN and 20.63 GBP, the London price in pence converted.