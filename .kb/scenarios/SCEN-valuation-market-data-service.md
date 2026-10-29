---
title: The owner ranks sources through the core's MarketDataService
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:54.456Z'
id: SCEN-valuation-market-data-service
type: scenario
---
Given an owner session on an allowed origin and the UI's view-plugin token, ListSourceRankings answers prices, FX rates and index readings, and SetSourceRanking stores an FX order; a source twice is INVALID_ARGUMENT, another caller PERMISSION_DENIED and an unknown session UNAUTHENTICATED.