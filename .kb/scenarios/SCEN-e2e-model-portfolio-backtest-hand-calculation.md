---
title: Turned on, a two-target basket over a synthetic series has the read API and the UI show its backtest across a yearly rebalance and each max drawdown as the hand calculation gives them
status: active
tags:
  - acceptance:p10-19
origin:
  kind: agent
  recorded_at: '2026-11-12T16:09:00.020Z'
id: SCEN-e2e-model-portfolio-backtest-hand-calculation
type: scenario
---
Given the data path's record, the institution fixture serving two synthetic index series, MODA from December 1, 2025 and MODB from December 3, each a round level that changes on a few weekdays, and the owner signed in to the built-in UI, when the owner saves a model portfolio of the two as index targets in PLN weighted 60% and 40% and turns it on, and an eodhd sync stores their levels, then the read API's GetModelPortfolio from December 3, 2025 and from March 6, 2026 equals the hand calculation: the backtest's index on every day across the yearly rebalance on December 31 to 12 decimal places, the model's max drawdown of 0.35 from January 20 to February 16 and each series' own, 0.25 from January 15 to February 2 for MODA and 0.55 from December 12 to February 16 for MODB, with peak and trough values, and the overview's performance chart draws the model as a second line over the performance's days, rebased on its first, with the three drawdowns in the table under it (D-53, D-54). Runs on both kind and k3d.