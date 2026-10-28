---
title: A market valuation is quantity times the end-of-day close in its own currency
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:57.303Z'
id: SCEN-valuation-market-value
type: scenario
---
Given an end-of-day close of 101.37 USD for an instrument and a holding of 12.5 units, when the holding is valued on the market basis, then the value is exactly 1267.125 USD with the price's date and source; given a price of a different instrument, then valuation fails.