---
title: A failed revaluation job keeps no values in River's log or its stored errors
status: active
verification_scope: integration
sourceFile: internal/valuation/postgres/joberror_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T15:15:41.763Z'
id: TEST-valuation-postgres-job-errors
type: test
---
Against a fresh PostgreSQL database, runs real River revaluation workers over a ledger whose logs fail with errors carrying an amount, an ISIN and an account number (a wrapped store error and an oversell), waits for the job to fail, and checks River's log lines with an error and the job's river_job.errors against a deny-list, and that both name the job. A worker returning the raw error fails it, as checked by hand.