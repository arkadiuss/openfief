---
title: '-route pins a model to hosts, labels each run, and the route is in the configuration hash'
status: active
sourceFile: tools/askbench/route_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T15:52:38.595Z'
id: TEST-askbench-route
type: test
---
Reads -route entries into runs of a model, one per entry, named by label; refuses a route with no provider or a repeated one; finds the structured flag by id or label; checks the routing member of a request on OpenRouter with a route, that the route changes the configuration hash and nothing else does, and that the report keys a model by its label with its route.
