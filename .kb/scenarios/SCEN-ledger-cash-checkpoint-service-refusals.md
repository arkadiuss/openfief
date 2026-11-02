---
title: The cash checkpoint service refuses what isn't the owner's well-formed entry
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:50:11.346Z'
id: SCEN-ledger-cash-checkpoint-service-refusals
type: scenario
---
Given the core's plugin port, when a caller that is not a view plugin, a read session or an unknown session calls the CashCheckpointService, or the authorizer doesn't let the owner change the account, or the owner sends a malformed date, currency or amount, names an account or checkpoint that is not stored, or books a checkpoint that already matches, then the call fails with permission_denied, unauthenticated, invalid_argument, not_found or failed_precondition and nothing is stored.