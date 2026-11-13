---
title: The plugin's manifest asks for what it needs
status: active
sourceFile: plugins/model-openai/manifest_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:04.919Z'
id: TEST-plugins-model-openai-manifest
type: test
---
Checks manifest.json: the model-provider capability, the host of baseUrl from config and no fixed host, the optional key and CA bundle Secrets, no data tier.