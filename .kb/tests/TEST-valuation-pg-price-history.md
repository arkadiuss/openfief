---
title: PostgreSQL returns an instrument's stored prices from every source in a range and the latest per source
status: active
sourceFile: internal/valuation/postgres/pricehistory_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T01:12:18.774Z'
id: TEST-valuation-pg-price-history
type: test
---
Runs under task test:pg on a fresh database with the valuation migrations. Saves synthetic prices of two listings of one ISIN from two sources, and of a bond, then checks PriceHistory returns the ISIN's prices in the range in listing, date and source order with nothing of another ISIN or outside the range, nothing for an unpriced ISIN, and LatestPrices one latest price per listing and source.