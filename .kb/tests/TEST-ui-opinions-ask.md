---
title: The opinions prompt, schema and thesis checks, the loop in each mode, the ineligible note, a thesis answer's history, th
status: active
sourceFile: internal/views/ui/askopinions_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: TEST-ui-opinions-ask
type: test
---
The opinions prompt, schema and thesis checks, the loop in each mode, the ineligible note, a thesis answer's history, th

The Ask box is a conversation between the box and a model: the opinions mode changes the prompt, the schema, the tiers the route is checked for and what the box does with a thesis, and a unit test with a scripted model is the only way to see all of them without a real model.

## Context

Asked for by the design session that handed over D-225 and D-229, 2026-11-21: with opinions on the model may say buy, sell or hold and suggest a thesis, and with them off, or on a public demo, it keeps declining (R-7).

## Source

> Opinion mode in the Ask box: prompt variant; "buy it for me" and "start the agent" declined.
