---
title: Opening cash is a deposit or withdrawal against equity that marks the account seeded
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:34.335Z'
id: SCEN-ledger-seed-opening-cash
type: scenario
---
Given an account's cash balance at go-live, when the connector books it with NewOpeningCash under an OpeningCashID key, then a positive balance is a deposit and a negative one a withdrawal, both against the equity ledger account in that currency, and IsSeeding recognizes it while an ordinary deposit is not. Assumes the connector keys opening cash with OpeningCashID.