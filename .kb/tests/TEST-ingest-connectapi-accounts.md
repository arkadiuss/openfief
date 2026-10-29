---
title: AccountService serves the owner's account entries to view plugins only
status: active
sourceFile: internal/ingest/connectapi/accounts_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:22.120Z'
id: TEST-ingest-connectapi-accounts
type: test
---
A generated Connect client against an httptest server of AccountService over in-memory fakes: create from a synthetic broker account, update and bind round-trip; a caller that is not a view plugin, no session and a read session are refused; a taken ID is ALREADY_EXISTS, an unknown sleeve or malformed instance INVALID_ARGUMENT, a broker account mapped elsewhere or never reported FAILED_PRECONDITION, an unknown account NOT_FOUND, a failing ledger INTERNAL, and no refused call binds anything.