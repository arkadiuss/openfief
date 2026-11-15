---
title: The chat client reads the length of the reasoning and not its text
status: active
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T12:26:15.057Z'
sourceFile: plugins/model-openai/openaichat/cached_test.go
verification_scope: unit
id: TEST-plugins-model-openai-reasoning-length
type: test
---
Completes against a fake server that puts reasoning in message.reasoning or message.reasoning_content, or in neither, and checks the reply's reasoning length and that its content is untouched.

## Context

Parsing check of the reasoning length.