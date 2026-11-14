---
title: The chat client speaks the OpenAI-compatible protocol
status: active
sourceFile: plugins/model-openai/openaichat/chat_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:04.529Z'
id: TEST-plugins-model-openai-chat
type: test
---
Sends requests to an httptest OpenAI-compatible server: path, key, tools, structured output and limits, the status classes, an unreachable provider and a key that cannot be read.