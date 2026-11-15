---
title: 'Re-grading a report offline: schema 2, schema 1 by replay, A19, kept and carried runs, the flags and the report''s stored cards'
status: active
sourceFile: tools/askbench/rescore_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Design session decisions (author''s delegate): a scorer version and offline re-scoring of the Ask box bench, with D-217, D-219, D-220 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T15:38:50.202Z'
id: TEST-askbench-rescore
type: test
---
Runs the command against an httptest server that plays the model and re-grades its report: a report of the current schema changes no check, a report stripped to schema 1 has its cards rebuilt from its calls and every check as graded, a hand-made A19 report of schema 1 passes the two real answers with the card's change and the one with both values and fails the percent rounded another way and the fall, a run whose replay differs is kept and counted, a check not stored fails the re-grading, the named cards are chosen only when the answer leaves no choice, the bench's report stores the cards and the output limit, the configuration holds no scorer, and -rescore needs -out, refuses the report's own directory and any flag of a run.