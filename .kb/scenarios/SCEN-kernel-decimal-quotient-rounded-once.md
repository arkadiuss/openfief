---
title: A quotient is rounded once, by the caller's mode
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:11:41.352Z'
id: SCEN-kernel-decimal-quotient-rounded-once
type: scenario
---
Given 1 divided by 8, when the quotient is rounded to 2 places, then half-up gives 0.13 and half-even gives 0.12; given 535 divided by 365 with half-up rounding to 2 places, then the result is 1.47; given any number divided by 0, then the division fails with a division-by-zero error and no value. Assumes the caller keeps the numerator exact and divides last.