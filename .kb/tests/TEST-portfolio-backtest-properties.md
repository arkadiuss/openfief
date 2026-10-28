---
title: The backtest of one target is its price, ignores price scale, and MaxDrawdown is the worst pair
status: active
sourceFile: internal/portfolio/backtest_property_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:20.508Z'
id: TEST-portfolio-backtest-properties
type: test
---
Seeded property tests (splitmix64, deterministic, no wall clock): a one-target model's index equals 100 x price / first price within 1e-10 on every day; multiplying one target's prices by a random factor, as a share split would, leaves a 70/30 model's index unchanged within 1e-9 across two year-end rebalances; MaxDrawdown's depth equals the brute-force largest fall over every pair of a day and a later one, and its trough is never before its peak.