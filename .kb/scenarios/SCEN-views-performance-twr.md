---
title: Deposits and withdrawals move the portfolio's value but not its time-weighted return
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:34.503Z'
id: SCEN-views-performance-twr
type: scenario
---
Given a synthetic portfolio valued every day, when money is deposited or withdrawn, then the day's factor is (value - net flow) / the previous day's value with the flow taken at the end of the day, so the index, 100 on the first day, moves only with prices; factors and their product are rounded half-even to 12 places. Assumes PLN reporting.