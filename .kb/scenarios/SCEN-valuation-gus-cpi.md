---
title: GUS's final year-on-year CPI of a year arrives as exact monthly fractions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:23.591Z'
id: SCEN-valuation-gus-cpi
type: scenario
---
Given GUS's public SDP API, when the CPI of a year is fetched, then the year-on-year index (same month a year before = 100) of each month comes back in month order as an exact fraction, 104.9 as 0.049, from the COICOP 1999 indicator up to 2025 and the COICOP 2018 one from 2026, stopping before the first month that is missing or only preliminary, estimated or forecast; a year with nothing published gives none.