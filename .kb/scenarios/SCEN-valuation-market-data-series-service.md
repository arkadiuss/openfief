---
title: The owner's series sources through the core
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:11.129Z'
id: SCEN-valuation-market-data-series-service
type: scenario
---
Given the core and the built-in UI's credentials with an owner session, when it sets the price gap threshold to 14 days, then the price ranking reports 14 and the FX ranking the default 7; a threshold of 0 and a source without a value of the series are refused as invalid argument, and the series list answers the owner.