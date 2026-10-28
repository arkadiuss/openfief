---
title: Tax converts at the NBP rate of the last business day before the event
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:33.881Z'
id: SCEN-pl-tax-fx-rate-date
type: scenario
---
Given income received on Monday 2026-03-09, when the NBP rate for Polish tax is chosen, then the cutoff is Sunday 2026-03-08, so the latest table on or before it, Friday's, applies; given income on Tuesday 2026-03-10, then Monday's table is the latest allowed and the event day's own table never applies.