---
title: Flagged balance assertions are listed per account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:15.149Z'
id: SCEN-ui-flagged-assertions
type: scenario
---
Given an account with a flagged balance assertion check, when the owner opens /assertions, then the check is listed under its account with its date, ledger account, reported and derived balances.