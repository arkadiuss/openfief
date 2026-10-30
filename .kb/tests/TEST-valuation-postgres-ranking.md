---
title: Every source's values are stored and reads return the ranked winners
status: active
sourceFile: internal/valuation/postgres/ranking_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:43.831Z'
id: TEST-valuation-postgres-ranking
type: test
---
PostgreSQL tests: two sources' prices of one day are both stored, Prices returns EODHD's by default and IBKR's on a day only it has, CrossCheckedPrices returns the lower-ranked one as a cross-check, the owner's ranking moves the winners and enqueues its revaluation in the same transaction, saving it again changes nothing and a source twice is refused; NBP's FX rate wins by default; the owner's first index reading source wins while both readings stay stored.