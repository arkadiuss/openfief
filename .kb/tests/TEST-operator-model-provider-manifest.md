---
title: The operator accepts the model-provider plugin's manifest and config schema
status: active
sourceFile: internal/plugins/operator/model_provider_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:05.676Z'
id: TEST-operator-model-provider-manifest
type: test
---
Parses the first-party plugin's manifest as the operator would and checks its schema against configs, and the config hosts the grant approves.