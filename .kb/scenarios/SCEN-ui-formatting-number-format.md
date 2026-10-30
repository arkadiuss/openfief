---
title: Amounts, chart axes and tooltips use the owner's number format
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:17.119Z'
id: SCEN-ui-formatting-number-format
type: scenario
---
Given the Polish default space-comma format, when any page renders 12345.67 EUR, then it reads 12 345,67 EUR and the chart data carries the separators; after the owner picks comma-dot, then 12,345.67 EUR, the digits and currency unchanged.