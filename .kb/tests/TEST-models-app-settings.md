---
title: The settings use cases are owner-only and default to off
status: active
sourceFile: internal/modelrouter/app/settings_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:27:59.704Z'
id: TEST-models-app-settings
type: test
---
Runs the settings use cases over a fake store and fake owner sessions: defaults, who may change what, the owner's changes and invalid input.