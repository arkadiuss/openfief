---
title: The core raises a transfer-suggested item for a suggested pair and clears it once confirmed
status: active
sourceFile: cmd/core/suggested_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T06:14:53.833Z'
id: TEST-core-suggested-transfers
type: test
---
Under task test:pg the core starts on a fresh PostgreSQL; two synthetic accounts get a withdrawal and a deposit of 1500 PLN two days apart through the ledger, one attention tick runs, and REST GET /v1/attention-items with a script token lists one open transfer-suggested item, subject the pair, group decisions, tied to both accounts; after the ledger records the owner's confirmation and another tick, nothing is open and one item is done.