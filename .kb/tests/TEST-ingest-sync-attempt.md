---
title: Sync error classes have fixed names, and a sync keeps failing only after three failures over fifteen minutes
status: active
sourceFile: internal/ingest/syncattempt_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-05T23:38:54.561Z'
id: TEST-ingest-sync-attempt
type: test
---
Table tests of ingest.SyncErrorClass's String and ParseSyncErrorClass for every class, the zero class and an unknown name, and of SyncStatus.KeepsFailing: never synced, synced, two failures over an hour, three within a minute, three just short of fifteen minutes, three over it, many over hours, and failures with no start. A false pass would be a rule that counts failures without their span; the cases on either side of the span catch it.