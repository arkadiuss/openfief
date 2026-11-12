---
title: test/e2e/modelcheck works out the model portfolio's backtest and max drawdowns by hand and agrees with portfolio.Backtest over the same series
status: active
sourceFile: test/e2e/modelcheck/modelcheck_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T16:08:57.916Z'
id: TEST-e2e-modelcheck
type: test
---
Runs offline under task test. Tests test/e2e/modelcheck, which works out by hand what the read API's model portfolio and the UI must show for the two synthetic index series. TestHandTableEqualsBacktest runs the core's own portfolio.Backtest and portfolio.MaxDrawdown over the series the fixture serves and requires every one of the 101 days, across the rebalance on December 31, and the basket's and each series' max drawdown, with peak and trough dates and values, to equal the hand-worked table. Line rebases the table on a first day, 100 on it, to 12 places half-even, from the backtest's start, from a later day, from before the backtest and refuses a day after its end; Aligned puts the line on a chart's days with gaps where the backtest has none; PerformanceChart adds it to the time-weighted return's line; Expected and Summarize agree on an answer shaped as the hand calculation, ignoring trailing zeros, and differ on a wrong day; DrawdownRows lists the three rows the page shows. A false pass would need the table and the code to share a mistake over the same levels.