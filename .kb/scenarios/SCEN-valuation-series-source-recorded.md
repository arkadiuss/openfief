---
title: A new ranking keeps every recorded series source
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:10.896Z'
id: SCEN-valuation-series-source-recorded
type: scenario
---
Given an instrument whose prices from two sources were first read while the first-party connector ranked first, when the owner ranks the other source first, then the instrument's prices still come from the first-party connector, nothing is revalued, and an instrument first read after the new ranking takes the newly first-ranked source.