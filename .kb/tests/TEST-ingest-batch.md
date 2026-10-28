---
title: Import batches and institutions validate and record a mapping's results
status: active
sourceFile: internal/ingest/batch_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:17:28.530Z'
id: TEST-ingest-batch
type: test
---
Table tests build institutions and import batches with synthetic IDs: an institution needs an ID and a name; NewBatch needs an ID, an institution, a source, a known ingest mode (opening or transactions) and a start date, and fails with ErrInvalidBatch otherwise. Complete records the counts of a mapping's transactions and balance assertions, the ledger's import summary and a copy of the ingest problems, checked by mutating the mapping's problems afterwards. Enum names and problem text are pinned.