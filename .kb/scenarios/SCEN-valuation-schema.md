---
title: The valuation schema stores inputs and valuations exactly
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:56:54.284Z'
id: SCEN-valuation-schema
type: scenario
---
Given a fresh database, when it is migrated twice, then the valuation tables and River's river_job exist; when prices, rates, readings, bond terms and a tax wrapper are saved, then they read back exactly; when valuations with a missing input are replaced from a date, then earlier rows stay and later ones are replaced.