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
Runs on a fresh PostgreSQL database with the valuation and River migrations (task test:pg). Migrating twice is a no-op and creates river.river_job and the valuation tables. Saving a price commits its dirty row and one River job; a batch whose second price violates a CHECK rolls back both. Three enqueues for acc-1 from 2026-03-10, 03-01 and 03-05 leave one job and from-date 2026-03-01. The ledger import hook enqueues on the import's own transaction through WithTx: commit keeps the jobs, rollback drops them, and a store without WithTx is refused. The bond basis enqueues only when it changes. Statuses report valued-through, pending, running and retry states set through SQL.