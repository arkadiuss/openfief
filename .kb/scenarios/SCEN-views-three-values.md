---
title: A reader gets every valuation and total with its three values, and an allocation of any of them
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:50.508Z'
id: SCEN-views-three-values
type: scenario
---
Given stored valuations with net and liquidation values, and an IKZE whose liquidation needs the unset marginal rate, when the read API lists valuations and totals, then each carries gross, net and liquidation or the reason one is missing, a total lacking a value leaves it unset, and the allocation of the liquidation value leaves out what lacks it, saying why.