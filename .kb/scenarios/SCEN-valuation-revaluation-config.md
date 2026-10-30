---
title: The revaluation workers start only with valid settings and a jurisdiction pack
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:05.512Z'
id: SCEN-valuation-revaluation-config
type: scenario
---
Given the revaluation workers' settings, when concurrency is below 1, the end of day has no time zone or an hour of 24, no date source or no jurisdiction pack is given, then building the client fails; with valid settings it revalues accounts end to end.