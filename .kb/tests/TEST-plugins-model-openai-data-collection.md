---
title: The plugin checks an entry's dataCollection and sends it with that entry's requests only
status: active
sourceFile: plugins/model-openai/datacollection_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T19:05:00.361Z'
id: TEST-plugins-model-openai-data-collection
type: test
---
Config tests: dataCollection is kept per entry on an OpenRouter base URL and refused on another server, for an unknown value and for a value in the wrong case; and a table test of the provider object each entry sends (pinned and denied, denied alone, with and without tools, an entry with none), with the shared client left unchanged.