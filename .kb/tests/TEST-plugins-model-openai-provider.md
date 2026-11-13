---
title: The plugin's service describes, lists and generates
status: active
sourceFile: plugins/model-openai/provider_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:03.467Z'
id: TEST-plugins-model-openai-provider
type: test
---
Calls the plugin's ModelProviderService over a fake provider: Describe, ListModels, Generate with tools and schemas, the requests it refuses before calling the provider, and provider failures mapped to Connect codes.