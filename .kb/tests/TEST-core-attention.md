---
title: The core serves the AttentionService and its checks raise items the read API lists
status: active
sourceFile: cmd/core/attention_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:45.803Z'
id: TEST-core-attention
type: test
---
Starts the core on a fresh PostgreSQL database: the plugin port serves the AttentionService to the owner's view and refuses another caller and an unknown session; a broker account recorded without a mapping raises one open item however often the checks run, the read API lists it over REST with a script token, and the outbox has nothing left unpublished after a tick.