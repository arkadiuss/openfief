---
title: The route check finds one problem per broken entry of a route in use
status: active
sourceFile: internal/modelrouter/app/routes_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T14:45:41.753Z'
id: TEST-models-app-routes
type: test
---
Runs RouteProblems over fake settings and provider instances: a missing instance, a plugin that is no model provider, two entries of one instance, a provider that is down or not running (none), a route that is not in use because the switch or the toggle is off (none), a core that has not found the plugins yet (an error, so nothing is cleared), and the item's subject and detail.