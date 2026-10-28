---
title: The allocation switches between its four dimensions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:11.201Z'
id: SCEN-ui-allocation-switch
type: scenario
---
Given the overview, when the owner picks currency in the allocation's switch, then htmx swaps in only the allocation for that dimension, read with that dimension; without htmx the browser gets the overview for it.