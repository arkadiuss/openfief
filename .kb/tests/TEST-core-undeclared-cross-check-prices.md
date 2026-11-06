---
title: Undeclared cross-check prices are dropped with an attention item through the real core and PostgreSQL
status: active
verification_scope: integration
sourceFile: cmd/core/undeclared_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:16.379Z'
id: TEST-core-undeclared-cross-check-prices
type: test
---
Runs two syncs of a connector instance through the core's connector command over a fresh PostgreSQL database, with a connector served over HTTP whose delivery carries a deposit and a cross-check price its Plugin resource doesn't declare. Each sync must import the deposit with one problem, and the attention store must then hold exactly one open item of kind cross-check-prices-undeclared for the instance, with the fixed detail and one change, so the second drop neither raises a second item nor reopens it. A core that stored the price, refused the delivery or raised nothing fails it.