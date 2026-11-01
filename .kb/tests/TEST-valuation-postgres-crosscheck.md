---
title: Cross-check prices are stored apart from prices in PostgreSQL and enqueue nothing
status: active
sourceFile: internal/valuation/postgres/crosscheck_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:07.771Z'
id: TEST-valuation-postgres-crosscheck
type: test
verification_scope: integration
---
Runs on a fresh database with every migration (task test:pg): saves synthetic eodhd prices, then IBKR cross-check prices from two sources for one day and one for a day without a price; saving one again replaces it; CrossCheckedPrices returns only the price of the day that has a cross-check price, and the cross-check prices in instrument, date and source order; no revaluation is pending and the stored valuation prices are unchanged; the table refuses a zero close.