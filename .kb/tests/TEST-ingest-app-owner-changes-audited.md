---
title: A ticker confirmed or corrected and a sync retried are entered before the change, and not made without the entry
status: active
sourceFile: internal/ingest/app/audit_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:43.307Z'
id: TEST-ingest-app-owner-changes-audited
type: test
---
Runs the Tickers and Retries use cases over fakes with a recording auditor: each enters one entry naming the instrument or the instance by ID, calls the change once, and with a log that is down calls no change; a retry naming no possible instance enters nothing.