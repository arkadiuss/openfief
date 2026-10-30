---
title: A lower-ranked price is checked against the winner
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:00.979Z'
id: SCEN-valuation-lower-ranked-cross-checks
type: scenario
---
Given prices-eu and ibkr-main prices of one day with prices-eu ranked first, the price checks compare ibkr-main's price with prices-eu's as a cross-check price; a cross-check price never becomes the price.