---
title: A sync hands the auditor what it wrote, and only a sync does
status: active
sourceFile: internal/ingest/app/connectors_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:47.763Z'
id: TEST-ingest-app-sync-audit
type: test
---
TestConnectorsSyncReportsOnlyWhatChangedState: a sync reports the transaction versions the ledger stored, the prices and bond terms that were new or different and the accounts they were stored in; one that re-sends what is stored reports an empty change; an upload is not audited as a sync.