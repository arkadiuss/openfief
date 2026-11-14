---
title: The router stops spending when the day's budget is used up
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:57.255Z'
id: SCEN-models-daily-budget
type: scenario
---
Given a budget of USD 0.000014 and a model whose calls cost USD 0.000007, when two calls are answered then a third is refused before any provider is called, with an audit entry of class budget and the end of the day, the day's total is the two costs, and the next day's first call is answered; without a cap nothing is read or recorded; a candidate priced in another currency is refused, and a usage that can't be read refuses while one that can't be written withholds the answer.