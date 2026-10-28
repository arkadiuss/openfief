---
title: The loader uses NBP rates, from the API or from the committed copy
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:51.932Z'
id: SCEN-demo-fx-rates
type: scenario
---
Given no network or a failing NBP API, when the loader needs FX rates, then it uses the committed NBP table A rates for USD and EUR from 14 days before go-live to today; given a working NBP API, then it fetches table A through the NBP adapter and keeps only the demo's currencies.