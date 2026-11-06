---
title: The Ask box's commands are parsed and matched by pure functions
status: active
sourceFile: internal/views/ui/ask_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T08:17:41.314Z'
id: TEST-ui-ask
type: test
---
Table tests of parseAsk over every command's phrasings and refusals, periodRange, matchTarget (exact, prefix, substring, kind), matchInstance, askSuggestions (each one understood), the known commands' examples, and returnPercent.