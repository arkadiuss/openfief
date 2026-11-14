---
title: The budget's pricing, cap and day, and the rolling question limits, are exact
status: active
sourceFile: internal/modelrouter/guards_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:50.776Z'
id: TEST-models-domain-guards
type: test
---
Table tests of modelrouter.CallCost (per-million prices in exact decimals, an entry with no price, another currency), Budget (the cap is reached exactly at the cap, no cap, NewBudget's refusals), BudgetDay (the core's zone, midnight, the end of a year), BudgetSpentError, QuestionLimits.Validate and DecideQuestion (rolling hour and day, the limit-th newest decides when it ends, both limits, forgotten and future questions).