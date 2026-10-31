---
title: The backtest starts at 100 when every target has data, converts at the latest rate and rebalances each year end
status: active
sourceFile: internal/portfolio/backtest_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:18.725Z'
id: TEST-portfolio-backtest
type: test
verification_scope: unit
---
Hand-computed table tests of Backtest over synthetic PLN and USD closes and rates: one target follows its close carried over missing days; two 50/50 targets worth 150 on Dec 31 rebalance to 3.75 and 7.5 units and are worth 112.5 on Jan 1, not 125; a last day on Dec 31 does not rebalance; a USD target is close times the latest rate and starts with its first rate; an index uses its own currency; the backtest starts on the latest first day and ends on the earliest last close; closes after through are ignored; the index keeps 12 places half-even; missing closes or rates and non-overlapping histories give no days; duplicate, mixed currency or non-positive closes and rates fail with ErrInvalidHistory. TestBacktestTargetHistories checks each target's own history in PLN.