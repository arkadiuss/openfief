---
title: Every source's values are stored and reads return the ranked winners
status: active
sourceFile: internal/valuation/postgres/ranking_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:43.831Z'
id: TEST-valuation-postgres-ranking
type: test
verification_scope: integration
---
PostgreSQL tests (task test:pg) of the market-data source store: synthetic prices of two connector instances for one instrument are both stored, but Prices returns only the first-party connector's, chosen from the default ranking on first need and recorded, so a day only the other instance priced is never filled from it; CrossCheckedPrices returns the other instance's prices as cross-checks; a new ranking keeps the recorded source and enqueues nothing, while an instrument first read afterwards takes the newly first-ranked source; the owner's switch revalues once and lists the source as the owner's choice with the sources that have values; FX rates and each index kind's readings come from their own source; gap thresholds are stored, replaced and refused out of range.