---
title: A whole run of the bench is added to the history and gates a route member
sourceFile: tools/askbench/history_run_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:21.788Z'
status: active
id: TEST-askbench-history-run
type: test
---
Runs the whole bench twice against the fake model with -history and -demo-route and checks the two runs are in the file with the report's hash and the day, the member's streak and that a failing member fails the run.