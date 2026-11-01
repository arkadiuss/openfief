---
title: An upload's mode follows how the account started, and opening cash goes with its opening upload only
status: active
sourceFile: internal/ingest/start_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:16.497Z'
id: TEST-ingest-start
type: test
---
Pure table tests of ingest.StartMode and ingest.CheckOpeningCash: a new account starts in opening mode or, overridden, in transactions mode; a seeded account takes transactions; an account that holds transactions but isn't seeded takes transactions and refuses an opening override with ErrStartedFromHistory; opening cash, even zero, is refused with a transactions upload and for a seeded account, and no cash goes with any upload.