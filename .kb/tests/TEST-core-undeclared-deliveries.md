---
title: Undeclared bond terms are dropped with an attention item through the real core and PostgreSQL
status: active
verification_scope: integration
sourceFile: cmd/core/deliveries_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T22:24:36.685Z'
id: TEST-core-undeclared-deliveries
type: test
---
Runs the core's connector command against a fresh PostgreSQL database and a connector served over HTTP whose delivery carries a deposit and one bond series' terms, with a fake cluster whose Plugin resource first says the grant doesn't approve bond terms. Two syncs must import the deposit with one problem, store no terms and leave exactly one open attention item for the instance; after the grant approves them, the next sync stores the terms with the instance as source and resolves the item.