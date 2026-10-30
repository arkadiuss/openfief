---
title: The owner confirms or rejects a pair through TransferService
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:14.034Z'
id: SCEN-ledger-transfer-service
type: scenario
---
Given a view plugin calling for an owner session, ConfirmTransfer and RejectTransfer record the decision and return the pair's link; a read session, a missing session or another caller is refused, and refused pairs map to invalid_argument, not_found or failed_precondition with nothing stored.