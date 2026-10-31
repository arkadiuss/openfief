---
title: BondLotService takes the owner's lot from a view plugin and maps every failure to its Connect code
status: active
sourceFile: internal/ingest/connectapi/service_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:48.665Z'
id: TEST-ingest-connectapi
type: test
verification_scope: integration
---
Serves the service over httptest with the use case on in-memory fakes and a view-plugin interceptor: a lot typed in percent reaches the connector as fractions (6.80% as 0.068) and answers the batch's counts. Refusals: not a view plugin and a read session are PERMISSION_DENIED, an unknown session UNAUTHENTICATED, malformed fields and a mismatched series INVALID_ARGUMENT naming the field, an account without a bond connector or a connector without bond lots FAILED_PRECONDITION, an unreachable connector UNAVAILABLE, a ledger failure INTERNAL without its cause.