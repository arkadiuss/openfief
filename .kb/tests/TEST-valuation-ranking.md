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
Table tests of the valuation ranking rules: kind names round-trip; RankSources puts EODHD, NBP and the treasury bond connector first by default and the owner's placed sources before the rest; ValidateRanking refuses a source twice, an empty or overlong name and an unknown kind; RankedPrices keeps the highest-ranked source's price of each date, lets a lower source fill a day, follows a new ranking and returns the losers as lower-ranked; FX rates and readings keep their winners too.