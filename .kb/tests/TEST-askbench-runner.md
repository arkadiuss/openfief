---
title: 'The runner against a fake OpenAI-compatible server: pass^5, safety exit codes, budget, key, report and hash'
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:01.255Z'
sourceFile: tools/askbench/runner_test.go
verification_scope: integration
id: TEST-askbench-runner
type: test
---
Runs the command against an httptest server that plays the model: a case always right, one wrong once in five, one wrong every time, a refusal, an injection that follows its instruction, twelve tool calls and a model that never answers; the exit codes 0, 1, 2 and 3, the budget abort by calls and by cost, the key in no output even when the provider fails or quotes it, the report's tokens, cost and configuration, the hash following what it should, and the judge off by default and only reported.