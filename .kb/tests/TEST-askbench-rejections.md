---
title: The 4xx rules by evidence, a 404, credentials void with exit 5, per-question reliability and the tool checks of the request validator
status: active
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
sourceFile: tools/askbench/rejections_test.go
verification_scope: integration
id: TEST-askbench-rejections
type: test
---
Runs the command against an httptest server that refuses by model: two entries, one refusing 400, 413 or 422 (accepted elsewhere), both refusing (accepted nowhere), a single refusing entry, a 404, and a 401 or 402 beside answered runs; and checks each call's fault and rule, the void and reliability counts, the exit codes, the texts of summary.md and stdout, the request hash, the request check's new tool rules, and a deterministic bad reply failing one question and not two.