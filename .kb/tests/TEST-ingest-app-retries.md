---
title: The owner's retry starts one sync at once, refused without the owner, while a sync runs, or within a minute
status: active
sourceFile: internal/ingest/app/retries_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T04:22:53.753Z'
id: TEST-ingest-app-retries
type: test
---
TestRetriesRefuseWithoutTheOwner drives Retries over fake owners, bindings and a recording trigger: the owner's retry starts one sync; a read session, a bound account the authorizer denies, or no running scheduler start none. TestRetriesRateLimit steps a clock: a retry refused because a sync runs doesn't count, a second within a minute is ErrRetryTooSoon, one a minute later starts. TestSchedulerRetry and TestSchedulerRetryRespectsAnotherClaim run the real scheduler over the in-memory store: an unlisted instance is refused, a retry runs in the background and is recorded as completed, and one while it syncs or under another core's claim is ErrSyncRunning.