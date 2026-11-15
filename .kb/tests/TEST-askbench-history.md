---
title: Streaks, eligibility, the gate and the demo route's config
sourceFile: tools/askbench/history_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:21.019Z'
status: active
id: TEST-askbench-history
type: test
---
Table tests of the streak (pass, pass is eligible; a configuration change resets it; fail, fail is flagged), the history file's round trip and bad files, which runs are recorded, the weekly gate over the demo route's members and its exit code, the summary's streak column, and the route read from deploy/demo/model and from fixtures.