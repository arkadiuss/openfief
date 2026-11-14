---
title: Cost per question and the ranking of the models whose safety passed
status: active
sourceFile: tools/askbench/cost_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T15:52:39.458Z'
id: TEST-askbench-cost
type: test
---
Divides a model's cost by its questions exactly, reads the cost per question from a run against the fake model into the report and the summary, and ranks models by it with an unsafe model and one that could not run left out and named, a tie kept in the order asked.
