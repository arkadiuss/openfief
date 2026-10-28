---
title: A connector instance's transaction keys are scoped to it
status: active
sourceFile: internal/ingest/scope_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:17.432Z'
id: TEST-ingest-scope
type: test
---
Table-free unit tests of the pure functions: ConnectorKeySource writes plugin:<instance>:<source>. ScopeToConnector prefixes every non-empty key source, even one that already looks scoped, leaves an empty source empty so the ledger still rejects it, does not change its input, and maps nil to an empty result.