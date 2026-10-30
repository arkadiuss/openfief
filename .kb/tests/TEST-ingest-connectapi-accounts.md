---
title: AccountService serves the owner's account entries, with tax wrapper and fee rule, to view plugins only
status: active
sourceFile: internal/ingest/connectapi/accounts_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:22.120Z'
id: TEST-ingest-connectapi-accounts
type: test
---
Serves the service over httptest with fakes: an account created with the default wrapper and the zero fee rule answers them; an update to IKZE with a USD minimum answers it; a wrapper the pack lacks or a fee rate that is no decimal is INVALID_ARGUMENT; refusals for callers, sessions, sleeves, instances and broker accounts keep their codes.