---
title: The read API serves the model portfolio from go-live while it is on and nothing while it is off
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:32:27.701Z'
id: SCEN-views-model-portfolio
type: scenario
---
Given a recorded portfolio and the model portfolio, when its read API method is called, then while the model is off it answers not found and names nothing; while it is on, it answers the targets, the backtest rebased to 100 on go-live or the asked from through the latest valued-through date or the asked to, the backtest's bounds, and the model's and each target's max drawdown.