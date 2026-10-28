---
title: Each input change revalues the accounts it affects from a fixed date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:55.055Z'
id: SCEN-valuation-revaluation-triggers
type: scenario
---
Given acc-1 holding an ETF and a bond and acc-2 holding another instrument, when an ETF price for 2026-03-04 is recorded, then only acc-1 is revalued from 2026-03-04; when an FX rate is recorded, then both are revalued from its date; when bond terms or a tax wrapper are recorded, then the holders or the account are revalued from their first transaction; when an import changes acc-1 from 2026-03-03, then acc-1 is revalued from that date.