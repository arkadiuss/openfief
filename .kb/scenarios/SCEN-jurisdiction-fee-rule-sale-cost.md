---
title: A sale's cost is the commission, at least the minimum, plus a conversion fee for a foreign instrument
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:51.283Z'
id: SCEN-jurisdiction-fee-rule-sale-cost
type: scenario
---
Given a fee rule of 0.29% with a 14 PLN minimum and a 0.5% conversion rate, when 10 000 PLN of a foreign instrument is sold, then it costs 79.00; 1 000 PLN of a domestic one costs the 14.00 minimum; the zero rule costs nothing; rates outside [0, 1) and a negative minimum are refused.