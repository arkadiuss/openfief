---
title: The FX rate for a date is the latest published on or before it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:29:50.669Z'
id: SCEN-valuation-fx-rate-lookup
type: scenario
---
Given NBP table A rates for USD on Friday 2026-03-06 and Monday 2026-03-09, when the rate for Sunday 2026-03-08 is looked up, then Friday's rate applies; when the rate for Monday is looked up, then Monday's applies; given no rate on or before the date, or two different rates for the chosen date, then the lookup fails instead of guessing.