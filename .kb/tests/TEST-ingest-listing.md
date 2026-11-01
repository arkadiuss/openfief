---
title: A confirmed ticker's listing settles the trading currency an opening lot's cost is booked in
status: active
sourceFile: internal/ingest/listing_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T16:47:51.572Z'
id: TEST-ingest-listing
type: test
---
Pure table tests of ingest.ListingOf, ingest.ParseTradingCurrency and TickerEntry.Transaction for opening lots priced by their open price: the confirmed listing's currency wins, a proposal of another ISIN doesn't count, a listing without one takes its exchange's, London quotes in pence, pence are divided by 100, costs round half to even to the currency's minor unit (none for JPY), and no known currency fails with ErrUnknownTradingCurrency.