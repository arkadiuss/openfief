---
title: The read API's ledger queries use only current versions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:29.898Z'
id: SCEN-ledger-app-queries
type: scenario
---
Given an account with buys, a sale, an FX conversion and a deposit corrected from another account into it, when its current transactions, holdings, lots and cash are asked for on a date, then they are derived from the current versions in that account only: the moved deposit counts in its new account and not in the old one, lots follow FIFO, and cash is reported per currency. A query for an account that is not stored is an unknown account error.