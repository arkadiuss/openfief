---
title: The model portfolio's report is computed from stored history only while it is on
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:23.860Z'
id: SCEN-portfolio-model-report
type: scenario
---
Given a model portfolio, when its report is asked for, then an off model reads nothing and reports nothing; an on model loads each instrument's prices by ISIN/symbol, each index's levels by the connector's name and the FX rates of their currencies through a date, and returns the backtest, the model's max drawdown over it and each target's over its own history.