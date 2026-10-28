---
title: The time-weighted return ignores deposit size, keeps cash flat and drifts within its bound
status: active
sourceFile: internal/views/performance_property_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:29.466Z'
id: TEST-views-performance-properties
type: test
---
Seeded property tests (splitmix64, 200 cases each): a portfolio fully invested in one ETF gets the same index day by day with no deposit and with an invested deposit of any size on any day; a portfolio of PLN cash alone stays at 100 whatever deposits and withdrawals move it; over 3700 days of random prices the chained index stays within 2e-12 of the growth per day of the exact 100 x last / first. Deterministic, no wall clock.