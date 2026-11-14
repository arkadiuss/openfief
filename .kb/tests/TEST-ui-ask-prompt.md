---
title: The system prompt's sentences, which of them are public, and its canary
status: active
sourceFile: internal/views/ui/askprompt_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T13:58:45.411Z'
id: TEST-ui-ask-prompt
type: test
---
Reads the system prompt of a question asked on a day: it holds the canary and ends with the day, its sentences in order are the prompt, one of them, what the box says of itself, is public, and the exported non-public sentences are the rest.
