---
title: A failed sync's error class comes from the known errors in its chain, and every attempt is recorded
status: active
sourceFile: internal/ingest/app/syncerror_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-05T23:38:55.750Z'
id: TEST-ingest-app-sync-errors
type: test
---
TestSyncErrorClassOf drives app.SyncErrorClassOf with wrapped errors: a deadline, a cancellation, an unreachable connector, the connector's own error, an untranslatable answer, an undeclared institution, an unbound account, ledger refusals and anything else. TestSchedulerRecordsAttempts runs the scheduler over the in-memory store with a failing then succeeding connector and checks both attempts' batch, mode, times, class, failures and next sync.