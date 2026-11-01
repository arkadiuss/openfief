---
title: TransferService confirms and rejects pairs for view plugins acting for the owner
status: active
sourceFile: internal/ledger/connectapi/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:06.894Z'
id: TEST-ledger-connectapi-transfers
type: test
verification_scope: integration
---
Serves TransferService over a fake decider behind a view-plugin interceptor: confirming and rejecting return the pair's link with its decision; a caller that is not a view plugin and a read session are permission_denied, no session unauthenticated, a pair that can't be a transfer invalid_argument, an unknown leg not_found, a taken leg failed_precondition and a store failure internal.