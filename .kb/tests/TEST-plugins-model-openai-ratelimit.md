---
title: The chat client retries a 429 once after Retry-After
status: active
origin:
  kind: agent
  ref: 'Design session decision: 429 handling, retried once honouring Retry-After (D-207), 2026-11-15'
  recorded_at: '2026-11-15T09:14:03.793Z'
sourceFile: plugins/model-openai/openaichat/ratelimit_test.go
verification_scope: integration
id: TEST-plugins-model-openai-ratelimit
type: test
---
Runs the chat client against an httptest server that answers 429 with each form of Retry-After, with a fake clock and a recording wait so that nothing sleeps: the wait, the capped wait, the default, the same request sent again, a second 429, other statuses never retried, and a wait past the deadline or cut short.
