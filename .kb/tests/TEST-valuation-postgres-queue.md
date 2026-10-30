---
title: Enqueueing commits with its data, coalesces per account, and reports status
status: active
sourceFile: internal/valuation/postgres/queue_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:55:50.422Z'
id: TEST-valuation-postgres-queue
type: test
---
PostgreSQL tests (task test:pg): inputs and imports enqueue in their own transaction, enqueues coalesce from the earliest date, a rules version enqueues only when it changed, and statuses report pending and running.