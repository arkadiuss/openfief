---
title: 'test/e2e/modelseries serves the model portfolio''s two synthetic index series: weekday closes from each first step to the end of the IBKR week, within a from and a to'
status: active
sourceFile: test/e2e/modelseries/modelseries_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T16:08:56.929Z'
id: TEST-e2e-modelseries
type: test
---
Runs offline under task test. Table tests of test/e2e/modelseries, the two synthetic index series the model portfolio's end-to-end check runs on: Find returns a series by its EODHD ticker and none for another; Closes gives a close on every weekday from the series' first step to the end of the IBKR week, none on a weekend, the new level on the day of a step, none before the first step, and only the days within a given from and to, a zero bound meaning none. The institution fixture and the hand calculation both import this package, so the levels the fixture serves are the levels the calculation assumes. A false pass would need a wrong level in the steps that the hand table also assumed, which the modelcheck test's comparison with portfolio.Backtest over the same closes would not catch; it would catch a table that disagrees with the code.