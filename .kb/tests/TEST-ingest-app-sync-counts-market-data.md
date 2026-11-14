---
title: A sync's change counts each kind of market data that was new or different
status: active
sourceFile: internal/ingest/app/connectors_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:02.178Z'
id: TEST-ingest-app-sync-counts-market-data
type: test
---
TestConnectorsSyncReportsOnlyWhatChangedState checks, over fake stores that answer a count for each kind, the change a sync hands the auditor: everything new, the same again (empty), only the transactions, and only one of prices, FX rates, index levels, index readings or bond terms new, so each kind has its own count.