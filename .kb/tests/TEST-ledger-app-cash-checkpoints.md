---
title: The ledger stores a cash checkpoint as an entered assertion, raises its item and books its difference
status: active
sourceFile: internal/ledger/app/cashcheckpoints_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:49:50.385Z'
id: TEST-ledger-app-cash-checkpoints
type: test
---
Over the in-memory store: the owner enters 1012.50 PLN against 1000 derived; the check is flagged and marked entered, the derived cash stays 1000, and the fake attention port gets the raise with subject acc-1/2026-03-31/PLN. Booking the difference deposits 12.50 on that date, the stored checkpoint matches, the cash is 1012.50 and the item clears; booking again finds nothing to book. Unknown accounts and checkpoints, a reported assertion, a failing attention port, a caller who is not the owner and an owner the authorizer doesn't let change the account are refused (D-158, D-163).