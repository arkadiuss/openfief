---
title: The Models page and its forms over a fake core
status: active
sourceFile: internal/views/ui/models_test.go
verification_scope: integration
origin:
  kind: agent
  ref: MS1 brief, 2026-11-13
  recorded_at: '2026-11-14T00:24:05.223Z'
id: TEST-ui-models
type: test
---
Tests of the built-in UI over a fake ModelSettingsService on an httptest server: the page with the switch, the toggles, the providers with their declared terms, the tier grid and the routes, a broken entry with its fix, who may see the page and its link, each form reaching the core with the owner's session, malformed forms never reaching it, and the core's refusals.