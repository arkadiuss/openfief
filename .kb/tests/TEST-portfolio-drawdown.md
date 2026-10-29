---
title: Max drawdown finds the worst fall with its dates, and a series rebases to 100 on a date
status: active
sourceFile: internal/portfolio/drawdown_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:22.269Z'
id: TEST-portfolio-drawdown
type: test
---
Table tests of MaxDrawdown on synthetic series: the deeper of two falls wins (130 to 65, 50%), a 50% fall from an early small peak beats a later 40% fall of more points, the earliest of equal falls is kept, the trough is the lowest point after the peak, a rising series and a single point have depth 0 with the first day as peak and trough, depth keeps 12 places, and a non-positive value or out-of-order dates fail. TestSince rebases from a day inside, before the first, after the last and on the last day.