---
title: Amounts and shares are formatted from exact decimals
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:21.227Z'
id: SCEN-ui-formatting
type: scenario
---
Given decimal strings from the read API, when the UI formats them, then it groups the integer digits in thousands and keeps every decimal place, computes shares half-even to one place with the kernel's decimals, and passes exact strings to the charts.