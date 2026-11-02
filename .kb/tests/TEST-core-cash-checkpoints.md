---
title: The core serves cash checkpoints on the plugin port and raises their attention items
status: active
sourceFile: cmd/core/cashcheckpoints_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T03:49:55.140Z'
id: TEST-core-cash-checkpoints
type: test
---
TestCashCheckpointService runs the real core on PostgreSQL: the owner creates an account, enters 250.00 PLN against nothing derived, the read API lists an open cash-checkpoint-mismatch item for it at once, booking the difference deposits 250 and resolves the item, and booking again, another caller and an unknown session are refused. TestLedgerAttention checks that the adapter passes the ledger's kind, which equals the attention context's, subject and detail through, and returns the store's failure (D-158, D-130).