---
title: The scheduler syncs due instances once at a time and retries the same batch with a backoff
status: active
sourceFile: internal/ingest/app/scheduler_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:00.395Z'
id: TEST-ingest-app-scheduler
type: test
verification_scope: unit
---
Drives Scheduler.Tick, SyncNow and Run with an injected clock, a fixed jitter, an in-memory SyncStore and a scripted connector, never sleeping. A never-synced instance syncs at once in opening mode, is not due before the interval plus jitter, then syncs in transactions mode with the stored watermark. Failures keep the batch open, retry it with the same ID after 5 then 10 minutes, and a success completes it and resets the failures. Only listed instances sync and an unlisted one keeps its watermark. A running sync is not started again and SyncNow then fails with ErrSyncRunning; SyncNow with another mode abandons the open batch; a timeout fails and records the attempt; a failing listing or a claim held elsewhere starts nothing; Run stops with its context. Bind checks the name and the account before storing.