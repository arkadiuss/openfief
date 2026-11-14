---
title: The core's audit of settings changes and its list of providers
status: active
sourceFile: cmd/core/modelsettings_test.go
verification_scope: unit
origin:
  kind: agent
  ref: MS1 brief, 2026-11-13
  recorded_at: '2026-11-14T00:24:07.256Z'
id: TEST-core-model-settings
type: test
---
Checks that a change of the models settings becomes a valid model_settings_changed entry of the audit log, that an unknown setting is refused, and that the providers can't be listed before discovery found the plugins.