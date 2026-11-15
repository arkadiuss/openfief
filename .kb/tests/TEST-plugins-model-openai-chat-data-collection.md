---
title: The chat client merges data_collection into the provider object on OpenRouter only
status: active
sourceFile: plugins/model-openai/openaichat/datacollection_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T19:04:59.469Z'
id: TEST-plugins-model-openai-chat-data-collection
type: test
---
A table test of the request body: deny and allow alone, with tools, with pinned hosts (all one provider object), an unknown value sends nothing, unset keeps the usual body, and another server is told nothing.