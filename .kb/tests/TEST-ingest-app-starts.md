---
title: An opening upload whose positions all wait starts the account at once
status: active
sourceFile: internal/ingest/app/starts_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T16:47:53.957Z'
id: TEST-ingest-app-starts
type: test
---
Runs app.Uploads over in-memory fakes of the connector, ledger, ticker store and account starts: an opening upload whose only entry is an opening lot waiting for its ticker records the account's go-live, holds the position, and the next upload of the same account is a transactions batch that books nothing dated on or before the go-live. A false pass would need the fake starts store to ignore the recorded date.