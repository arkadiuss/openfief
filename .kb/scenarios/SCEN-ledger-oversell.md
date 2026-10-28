---
title: Selling more than is held is an error
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:07.706Z'
id: SCEN-ledger-oversell
type: scenario
---
Given an account that holds 3 units of an instrument, when a sale of 4 units is derived into lots, then derivation fails with an oversell error naming the account, instrument, date, the sale's key, 4 requested and 3 available, and no lot is consumed.