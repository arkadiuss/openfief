---
title: The CashCheckpointService maps the owner's checkpoint and booking and their refusals to Connect
status: active
sourceFile: internal/ledger/connectapi/cashcheckpoints_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T03:49:52.777Z'
id: TEST-ledger-connectapi-cash-checkpoints
type: test
---
Starts the CashCheckpointService on an httptest server over a fake ledger: recording a checkpoint answers the cash shown and derived, booking answers the key, kind and amount; another caller, no session or a read session, a malformed date, currency or amount, an unknown account or checkpoint, nothing to book and a store failure each map to their Connect code, so the UI can show the owner why (D-158, D-79).