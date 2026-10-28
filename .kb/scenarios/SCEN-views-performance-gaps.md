---
title: The performance series stops before the first day that lacks an input
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:41.627Z'
id: SCEN-views-performance-gaps
type: scenario
---
Given a portfolio whose valuation lacks a price on a day, has no row or a stale quantity for a holding, is valued only through an earlier date, or holds cash in a currency without an FX rate, when the series is computed, then it ends on the day before and names the gap, unvalued or no FX rate, instead of drawing a drop.