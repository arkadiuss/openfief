---
title: A long FX date range is fetched in deterministic chunks within the provider's limit
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:54:14.886Z'
id: SCEN-valuation-fx-range-splitting
type: scenario
---
Given a rate provider that serves at most 93 days per request, when rates from 2025-01-01 to 2025-04-04 are fetched, then it is asked for 2025-01-01 to 2025-04-03 and then 2025-04-04 to 2025-04-04, in that order, and the combined rates come back sorted by date then currency; given a range that ends before it starts, then fetching fails with an invalid range error before any request.