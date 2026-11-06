---
title: A connector sync that keeps failing raises one attention item, which clears after a success
status: active
sourceFile: internal/attention/app/syncs_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:34.364Z'
id: TEST-attention-app-failing-syncs
type: test
---
TestFailingSyncs runs app.FailingSyncs over fixed sync statuses: a synced instance, one that failed once and one that keeps failing, and a failing store, and checks only the last is a condition, its subject the instance and its detail when the failures began. TestFailingSyncItem sweeps the check through a run of failures with attention.Reconcile: not raised after two, raised after three, kept acknowledged while the run goes on, cleared after a success and reopened for a new run.