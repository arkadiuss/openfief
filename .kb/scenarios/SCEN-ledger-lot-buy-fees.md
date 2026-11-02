---
title: A buy's fees go to its lot and split with its units
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:12.637Z'
id: SCEN-ledger-lot-buy-fees
type: scenario
---
Given a buy of one holding with a fee, when the lots are derived, then its lot carries the fee, and a sale of part of it takes the fee's share of the units rounded half-even to two places while the lot keeps the exact rest, so 1.00 USD and 4.00 PLN over three units split as 0.33, 0.34, 0.33 and 1.33, 1.34, 1.33.