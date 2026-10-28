---
title: Sync states, claims, batch history and connector bindings round-trip through PostgreSQL
status: active
sourceFile: internal/ingest/postgres/store_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:02.528Z'
id: TEST-ingest-postgres
type: test
---
Runs under task test:pg on a fresh database migrated twice. A first claim returns an empty state and a second one before the lease ends fails with ErrSyncRunning; after StartBatch and FailSync the next claim returns one failure and the open batch with its mode; CompleteSync stores the watermark, marks the instance seeded, completes the batch with its counts and attempts, and a completed batch can neither complete nor fail again. A claim lapses when its lease expires. Starting another batch abandons the open one and a batch ID can't be reused. Bindings are idempotent, listed in ID order, per instance, and the table refuses an invalid instance name.