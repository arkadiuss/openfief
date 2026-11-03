---
title: The allocation converts cash to the reporting currency
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:51.230Z'
id: SCEN-views-allocation-inputs-reporting
type: scenario
---
Given holdings with stored valuations and cash in several currencies, when the allocation is asked for a date, then each holding counts at its stored valuation, one without counts as unvalued, and each cash balance is converted to the reporting currency at the latest rate into it on or before the date, failing without one (D-30, D-180).