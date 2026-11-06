---
title: The activity's batches and a sync's attempts name viewable accounts with their counts
status: active
sourceFile: internal/views/app/batchcounts_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:12.756Z'
id: TEST-views-app-batch-counts
type: test
---
Runs Views.Activity and Views.SyncAttempts over an in-memory sync history with counts per account, as the owner and as a viewer of one account: the upload's totals and accounts, and the accounts an attempt's batch booked into.