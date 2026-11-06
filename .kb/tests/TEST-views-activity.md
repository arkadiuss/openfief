---
title: The recent activity merges its three sources newest first and pages without gaps
status: active
sourceFile: internal/views/activity_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:19.570Z'
id: TEST-views-activity
type: test
---
Table tests of views.MergeActivity over synthetic batches, failed attempts and ledger changes: newest first, ties in source order, hidden changes paged past, a source that runs out while it may hold more stopping the page; reading every page with the returned cursor yields each entry once. Also the cursor's encode and parse round trip with malformed tokens refused, and the kind each batch, attempt and change maps to.