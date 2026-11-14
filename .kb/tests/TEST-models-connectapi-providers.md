---
title: ListModelProviders and the broken entries over Connect
status: active
sourceFile: internal/modelrouter/connectapi/providers_test.go
verification_scope: integration
origin:
  kind: agent
  ref: MS1 brief, 2026-11-13
  recorded_at: '2026-11-14T00:24:06.619Z'
id: TEST-models-connectapi-providers
type: test
---
Calls the ModelSettingsService over an httptest server with fake plugins: the providers with their terms and models, an unavailable one with none, the owner's session alone, and the broken route entries in the settings.