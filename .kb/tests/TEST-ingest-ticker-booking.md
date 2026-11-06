---
title: Ticker holdings are asserted once confirmed and a booking books every instance's rows
status: active
sourceFile: internal/ingest/tickerbooking_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T06:59:48.246Z'
id: TEST-ingest-ticker-booking
type: test
---
Table tests over synthetic tickers: a ticker assertion validates its symbol, date and quantity and becomes the balance assertion of the confirmed instrument; BookTickerAssertions asserts confirmed tickers' holdings and returns the rest to wait; BookedTickerRows keeps a delivery's booked entries and holdings for a correction; TickerBooking.Ledger books two instances' rows with the listing, keys scoped to each instance, and for a correction to another ISIN returns the old instrument's holdings to withdraw, none for a currency correction. A false pass would need a correction to book under the old ISIN or a holding to be asserted before its ticker is confirmed (D-153, D-155).