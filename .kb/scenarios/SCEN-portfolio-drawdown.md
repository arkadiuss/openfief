---
title: The max drawdown is the worst fall from a peak to a later trough, with both dates
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:29.420Z'
id: SCEN-portfolio-drawdown
type: scenario
---
Given a positive series in date order, when its max drawdown is computed, then it is the largest fall from a day's value to a later day's as a fraction of the first, with the peak's and trough's dates and values, the earliest of equal falls, depth rounded half-even to 12 places, and depth 0 on the first day for a series that never falls.