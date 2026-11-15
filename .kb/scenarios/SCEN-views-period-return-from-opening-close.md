---
title: A period's return counts its first day in full
status: active
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: D-223, a period''s return runs from the close of the day before its first day, 2026-11-15'
  recorded_at: '2026-11-15T18:12:31.460Z'
id: SCEN-views-period-return-from-opening-close
type: scenario
---
Given an account valued every day through 2025-12-31 with a deposit inside the year, when the portfolio's change is asked for 2025 and for November 2025, then each period's opening day is the day before its first day (2024-12-31 and 2025-10-31), its return runs from the end of that day to the end of its last day, and the first day's own move is in it.