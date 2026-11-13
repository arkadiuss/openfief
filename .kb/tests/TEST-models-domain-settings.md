---
title: Settings.HasModel is true only with the switch, the toggle and a route
status: active
sourceFile: internal/modelrouter/settings_test.go
verification_scope: unit
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:31.581Z'
id: TEST-models-domain-settings
type: test
---
Table test of the switch, the feature toggles and the routes: it exists so the Ask box knows whether a question beyond its commands has a model to go to, and falls back to the commands alone otherwise.