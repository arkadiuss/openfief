---
title: Ticker entries book with a confirmed ISIN and otherwise wait as ingest problems
status: active
sourceFile: internal/ingest/ticker_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:21.298Z'
id: TEST-ingest-ticker
type: test
---
Pure table tests of the ingest domain: a buy, sell and dividend ticker entry each book the transaction of its kind with the instrument made of the confirmed ISIN and the ticker's symbol, a dividend with units or another kind is refused; BookTickerEntries books confirmed entries and returns the others with one waiting problem each without changing its input; tickers validate their symbol and country; and ResolveBrokerAccounts gives ticker entries of a mapped broker account its OpenFief account and key source.