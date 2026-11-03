---
title: An account is stored when valid, and imports need a stored account
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:14.115Z'
id: SCEN-ledger-app-accounts
type: scenario
---
Given the ledger's use cases, when a valid account is saved it is stored and replaces the one with its ID, an invalid one is refused and nothing changes, and an import of a transaction into an account that is not stored is refused as an unknown account.