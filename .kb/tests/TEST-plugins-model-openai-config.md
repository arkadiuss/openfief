---
title: The plugin checks its provider instance's config, an entry's pinned hosts and reasoning included
status: active
sourceFile: plugins/model-openai/config_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:04.185Z'
id: TEST-plugins-model-openai-config
type: test
---
Table tests of the config: locations, retention, timeouts, entries, prices and duplicate ids; and the OpenRouter options of an entry: hosts and reasoning are kept per entry on an OpenRouter base URL, and refused on another server, for a host that is no slug, listed twice or too long, for too many hosts and for an unknown reasoning value.