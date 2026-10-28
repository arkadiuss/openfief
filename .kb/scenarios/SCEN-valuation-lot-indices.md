---
title: A bond lot takes the index readings of the periods it has started
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:50.597Z'
id: SCEN-valuation-lot-indices
type: scenario
---
Given CPI readings for January 2025 and 2026 and a reference-rate reading for April 2026, when an EDO lot bought on 2024-03-15 is valued on 2025-03-15, then period 2 takes the January 2025 CPI; when a ROR lot bought on 2026-03-10 is valued on 2026-04-10, then period 2 takes the April 2026 reading; when a TOS lot is valued, then it needs no reading.