---
title: The database keeps the log append-only and imports serialize versions
status: active
sourceFile: internal/ledger/postgres/appendonly_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:28:52.674Z'
id: TEST-ledger-pg-log
type: test
verification_scope: integration
---
On a fresh database: appending a stored version again fails with ErrVersionConflict, and version 3 without version 2 fails with SQLSTATE OFA02, both leaving the versions unchanged. UPDATE, DELETE and TRUNCATE of transaction versions and postings, and inserting a posting for a stored version, each fail with SQLSTATE OFA01 from the triggers, and the log reads back unchanged. A posting changed with the trigger switched off by the owner makes Log and Versions fail with ErrCorruptLog through the fingerprint check. app.Ledger over the store imports a deposit and buy with a matched and a flagged assertion, re-imports them unchanged, and eight goroutines importing one correction at once append exactly one version 2, with the earliest affected date reported by the import that appended it and the stored checks re-run.