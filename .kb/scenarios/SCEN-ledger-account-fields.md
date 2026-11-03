---
title: An account needs an ID, an institution and a name, and nothing more
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:09.498Z'
id: SCEN-ledger-account-fields
type: scenario
---
Given an account with an ID, an institution and a name, when it is validated, then it is valid with no sleeve and no payout setting; given one without its ID, institution or name, then it is rejected as an invalid account.