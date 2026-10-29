---
title: The backtest starts at 100 when every target has data and rebalances at each year end
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:27.685Z'
id: SCEN-portfolio-backtest-rebalance
type: scenario
---
Given a model whose targets all have histories, when it is backtested, then it starts at 100 on the first day every target has data, held at the target weights, runs to the last day every target has data, and at the end of each December 31 before that rebalances to the target weights; a target without history gives no days.