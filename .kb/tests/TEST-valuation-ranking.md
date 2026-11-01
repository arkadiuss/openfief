---
title: Sources are ranked by the owner's order then by default, and the highest-ranked value of each key wins
status: active
sourceFile: internal/valuation/ranking_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:39.526Z'
id: TEST-valuation-ranking
type: test
verification_scope: unit
---
Table tests of internal/valuation/ranking.go: every kind of market data round-trips through its name and an unknown one is refused; RankSources puts the owner's placed sources first in the owner's order and every other known source after them, the first-party institution of the kind first (EODHD for prices, NBP for FX rates, the treasury bond connector for index readings), then by name; ValidateRanking refuses a source named twice, an empty or overlong name and an unknown kind. The ranking only chooses a series' source on first need (D-94), which TEST-valuation-series covers.