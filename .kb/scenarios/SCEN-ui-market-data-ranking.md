---
title: The owner moves a source on the settings page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:56.661Z'
id: SCEN-ui-market-data-ranking
type: scenario
---
Given prices-eu ranked above ibkr-flex-a, when the owner presses Up beside ibkr-flex-a, the UI sends ibkr-flex-a then prices-eu to SetSourceRanking and returns to the settings; Back to the default order sends an empty order; the core's refusal shows its reason.