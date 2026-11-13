---
title: The chat client speaks the OpenAI-compatible protocol
status: active
sourceFile: plugins/model-openai/chat_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:02.715Z'
id: TEST-plugins-model-openai-chat
type: test
---
Sends requests to an httptest OpenAI-compatible server: path, key, tools, structured output and limits, the status classes, an unreachable provider and a key that cannot be read.