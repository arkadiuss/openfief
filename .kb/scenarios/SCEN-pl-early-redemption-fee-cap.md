---
title: The early-redemption fee never exceeds accrued interest
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:27.466Z'
id: SCEN-pl-early-redemption-fee-cap
type: scenario
---
Given an EDO bond with a 2.00 PLN fee that has accrued 0.51 PLN of interest, when it is valued as if redeemed today, then the fee charged is 0.51 PLN and the bond returns its 100 PLN nominal; given a COI bond in its second year with 0.14 PLN accrued, then the full 2.00 PLN fee is charged; given any bond at maturity, then no fee is charged.