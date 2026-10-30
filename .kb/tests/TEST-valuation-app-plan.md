---
title: Which accounts each input change revalues, and recording inputs
status: active
sourceFile: internal/valuation/app/plan_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:27.181Z'
id: TEST-valuation-app-plan
type: test
---
Planner and InputRecorder tests over fakes: prices, rates, readings, terms, a tax wrapper and a fee rule revalue their accounts by fixed rules, the owner's marginal rate every account, the number format nothing; invalid input, an unknown wrapper, a 100% fee rate, a rate of 32 and a format with one separator twice are refused storing nothing; ApplyRules records pl/1 and revalues every account from its start.