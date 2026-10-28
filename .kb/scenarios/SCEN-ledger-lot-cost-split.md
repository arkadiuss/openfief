---
title: Partly consuming a lot splits its cost without creating or losing money
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:05.616Z'
id: SCEN-ledger-lot-cost-split
type: scenario
---
Given a lot of 3 units that cost 100.00 USD, when its units are sold one at a time, then the consumed costs are 33.33, 33.34 and 33.33 USD (half-even to 2 places, the last unit taking what is left), and at every step the consumed and remaining cost add up to 100.00 USD.