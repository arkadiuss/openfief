---
title: Exact arithmetic, rounding only on request
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:27.691Z'
id: SCEN-kernel-decimal-exact
type: scenario
---
Given 0.1 and 0.2, when they are added, then the result is exactly 0.3. Given a price of 101.37 USD and a quantity of 3.125, when they are multiplied, then the value is exactly 316.78125 USD, and it becomes 316.78 USD only when the caller rounds it to 2 places with half-even rounding. Assumes no float64 on the path from text to result.