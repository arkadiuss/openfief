---
title: The chat client reads the cached tokens a host reports
status: active
sourceFile: plugins/model-openai/openaichat/cached_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'Live bench finding on R1 (gpt-oss-120b on deepinfra: two model calls, 1694 input tokens, a greeting), 2026-11-14'
  recorded_at: '2026-11-14T23:20:13.308Z'
id: TEST-openaichat-cached-tokens
type: test
---
Checks that Complete reads cached tokens from usage.prompt_tokens_details.cached_tokens, else from usage.cached_tokens, and reports 0 for a null detail or none.