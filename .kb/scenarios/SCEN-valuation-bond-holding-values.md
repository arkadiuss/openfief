---
title: A bond holding sums its lots' three values in its account's wrapper
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:57.508Z'
id: SCEN-valuation-bond-holding-values
type: scenario
---
Given an EDO holding of two lots, when it is revalued in taxable, IKE, IKZE or OKI, then its gross value is the lots' gross accrued values summed, and its net and liquidation values the sums of what the pack computes per lot from interest, discount and fee; a missing index reading stores a gap.