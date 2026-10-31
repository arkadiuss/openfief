---
title: The core's plugin port serves TransferService to the owner's view only
status: active
sourceFile: cmd/core/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:32:31.169Z'
id: TEST-core-transfers
type: test
verification_scope: integration
---
Over a real core with PostgreSQL (task test:pg) and an owner session, ConfirmTransfer from the UI's plugin token reaches the ledger, which answers not found for legs it doesn't hold, while a caller that is not a view plugin is permission_denied and an unknown session unauthenticated.