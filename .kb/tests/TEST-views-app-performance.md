---
title: The performance query resolves its range, statuses and inputs over the ports
status: active
sourceFile: internal/views/app/performance_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:32.031Z'
id: TEST-views-app-performance
type: test
verification_scope: unit
---
Over in-memory fakes of the ledger, valuations, statuses and FX rates with a synthetic two-account portfolio: the series runs from go-live to the latest valued-through date by default, a from before go-live starts at go-live, a later from restarts the index at 100, an account with transactions that was never valued stops the series at once, no transactions give no days, a pending revaluation is reported, and a reversed range or a failing store fails the call.