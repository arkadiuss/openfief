---
title: An instrument's transactions, listings, holdings, prices and exemption classes are picked out purely
status: active
sourceFile: internal/views/instrument_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:20.775Z'
id: TEST-views-instrument
type: test
---
Table tests of views.InstrumentTransactions and InstrumentListings over synthetic buys of two listings of one ISIN and another ISIN, ListingOf over good and malformed listings, HoldingsOf with lots of each listing, AssetOf for a bond and an ETF, InstrumentClasses with the Polish pack's and the owner's class, and MarkPrimary marking only the prices of each listing's own source.