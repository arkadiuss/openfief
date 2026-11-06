---
title: A batch's counts per account from its mapping and the ledger
status: active
sourceFile: internal/ingest/counts_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:15.336Z'
id: TEST-ingest-count-per-account
type: test
---
Checks ingest.CountPerAccount over a synthetic mapping of three accounts and the ledger's counts: transactions per account from the mapping, the ledger's counts merged, an assertion-only account kept, an unmapped one dropped.