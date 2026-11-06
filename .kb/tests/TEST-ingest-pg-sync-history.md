---
title: PostgreSQL keeps every sync attempt append-only, sums each instance's status up and pages completed batches
status: active
sourceFile: internal/ingest/postgres/history_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-05T23:38:58.130Z'
id: TEST-ingest-pg-sync-history
type: test
---
Runs under task test:pg on a fresh database with the ingest migrations. Drives xtb-main through two failures, a success and three failures over twenty minutes, and nbp through one success, then checks the attempts newest first with their classes, failures and next syncs, paging failed ones after a cursor, each instance's status (failures since the latest success, failing since, latest class, KeepsFailing), that a DELETE is refused, and completed batches of every source newest first with a cursor.