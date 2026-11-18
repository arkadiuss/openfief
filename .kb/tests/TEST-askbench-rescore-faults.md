---
title: Re-scoring counts reliability per question from stored call failures and derives the faults an older report did not store
status: active
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
sourceFile: tools/askbench/rescorefaults_test.go
verification_scope: integration
id: TEST-askbench-rescore-faults
type: test
---
Re-scores a report written before scorer version 3 with a timeout, a 402, a 400 and a 404 among its failed calls, and checks the faults it derives, the void runs, the per-question reliability, the counts of derived and undecided faults in the report, summary.md and the console text, the exit code 5 and the unchanged configuration hash; and that re-scoring a report of this build leaves its reliability as it was.