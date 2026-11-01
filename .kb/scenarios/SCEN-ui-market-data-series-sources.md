---
title: The owner sees and switches a series' source
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:18.676Z'
id: SCEN-ui-market-data-series-sources
type: scenario
---
Given the owner on the settings page, when it shows market data, then each kind has its gap threshold field and every series lists its sources with the current one chosen and whether it came from the ranking; saving another source or a threshold hands it to the core, and the core's refusal is shown with its reason.