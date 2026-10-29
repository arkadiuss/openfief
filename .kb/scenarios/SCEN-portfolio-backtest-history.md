---
title: Each target's history is its latest close times the latest rate into the reporting currency
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:25.845Z'
id: SCEN-portfolio-backtest-history
type: scenario
---
Given a target's end-of-day closes or levels and FX rates, when its history is built, then each day from the first with a close and, in a foreign currency, a rate, to its last close on or before through, is valued at the latest close times the latest rate; a target without closes or without a rate before its last close has no history and says which is missing.