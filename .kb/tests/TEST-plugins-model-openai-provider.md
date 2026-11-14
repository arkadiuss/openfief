---
title: The plugin's service describes, lists and generates, each entry with its own pinned hosts and reasoning
status: active
sourceFile: plugins/model-openai/provider_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:03.467Z'
id: TEST-plugins-model-openai-provider
type: test
---
Calls the plugin's ModelProviderService over a fake provider: Describe, ListModels, Generate with tools and schemas, the requests it refuses before calling the provider, and provider failures mapped to Connect codes. A table test checks the body each entry of an OpenRouter instance sends: its own hosts with no fallback and require_parameters, its own reasoning, nothing for an entry with neither, nothing at another server, and the shared client unchanged.