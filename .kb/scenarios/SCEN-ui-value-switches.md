---
title: The allocation and value-over-time chart switch between the three values, starting on gross
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:14.779Z'
id: SCEN-ui-value-switches
type: scenario
---
Given the overview with no value picked, when it loads, then both switches start on gross; when the owner picks liquidation for the allocation and net or all three for the value chart, then htmx swaps in that value, and the value chart skips days without it, apart from the performance chart.