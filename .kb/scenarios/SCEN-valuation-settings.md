---
title: Revaluation settings are configurable and checked
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:56:45.334Z'
id: SCEN-valuation-settings
type: scenario
---
Given a daily end-of-day time of 23:30 in Europe/Warsaw, when the next run is asked for at noon, at 23:30, from a UTC instant, or on the day summer time starts, then it is 23:30 Warsaw time that day or the next; when concurrency is zero, the time zone missing, the hour 24, the date source missing or the bond basis market, then creating the client fails.