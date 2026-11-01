---
title: A River worker revalues an account end to end, and the settings are checked
status: active
sourceFile: internal/valuation/postgres/worker_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:56:06.973Z'
id: TEST-valuation-postgres-worker
type: test
verification_scope: integration
---
Runs a real River client against a fresh database with an in-memory ledger log (acc-1 buys 10 synthetic ETF units on 2026-03-02) and waits on River's completion events with a 30 s deadline, no sleeps. Recording an FX rate and a 2026-03-03 price enqueues one coalesced job; the worker stores a missing-price row for 2026-03-02 and 4000 PLN for later days, valued through 2026-03-01, nothing left pending. The missing price's own job closes the gap; the end-of-day job enqueues and runs another. DailyAt picks the next run in its time zone across DST, TodayIn reads the date in a zone, and invalid concurrency, time zone, hour, date source or a market bond basis are rejected.