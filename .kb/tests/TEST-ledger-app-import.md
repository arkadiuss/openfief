---
title: Import appends versions idempotently and survives version races
status: active
sourceFile: internal/ledger/app/import_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:11.066Z'
id: TEST-ledger-app-import
type: test
verification_scope: unit
---
Table tests drive app.Ledger.Import over an in-memory store that, like the Postgres one, refuses a repeated or skipped version. New keys become version 1; re-importing the same rows, or a row whose postings are reordered and written as 1.00 instead of 1, appends nothing and counts as unchanged; a changed row becomes version 2 while version 1 stays byte-for-byte; changing it back is version 3; an unknown account or an undated row fails the whole import and writes nothing. A race hook appends for another import between read and write: when it appended the same content the retry counts it unchanged, when it appended other content the retry appends the next version, and when it wins every time the import fails with a *VersionConflictError wrapping ErrVersionConflict after MaxAppendAttempts. Eight goroutines importing one correction at once append exactly one version 2. The stored log is compared in full with go-cmp, so a duplicate or lost version fails.