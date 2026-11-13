---
title: The tool loop shows exactly the cards used, tags every message with a tier, withholds each result the route may not receive and is bounded
status: active
sourceFile: internal/views/ui/askmodel_test.go
verification_scope: unit
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:33.846Z'
id: TEST-ui-ask-model-loop
type: test
---
Runs the loop over a scripted router port: cards shown against cards sent, tiers per message, per-result tier refusal, limits, failures of calls and tools, fallbacks, and no text in the log.