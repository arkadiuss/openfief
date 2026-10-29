---
title: StatementService takes an owner's upload stream from a view plugin and answers each failure with its own code
status: active
sourceFile: internal/ingest/connectapi/statements_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:02.978Z'
id: TEST-ingest-connectapi-statements
type: test
---
Serves the service over httptest (HTTP/1.1, as the UI calls it) with the use case on fakes and a streaming view-plugin check: a synthetic XML statement in 64 KiB chunks reaches the connector whole as statement.xml, text/xml, transactions, and the answer carries the counts and 100 of 105 problems. Refusals: not a view plugin and a read session PERMISSION_DENIED, an unknown session UNAUTHENTICATED, no header, a second header, no mode and an empty file INVALID_ARGUMENT, a file over the limit and a message over 1 MiB RESOURCE_EXHAUSTED, an account without a connector, no file support and an already seeded account FAILED_PRECONDITION, a refusal INVALID_ARGUMENT with its reason, an unreachable connector UNAVAILABLE, a ledger failure INTERNAL without its cause.