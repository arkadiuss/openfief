---
title: EODHD listings are picked by ISIN on supported exchanges and quoted in their currency
status: active
sourceFile: internal/valuation/eodhd/listing_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T21:36:56.873Z'
id: TEST-valuation-eodhd-listing
type: test
---
Pure table tests of the adapter's listing rules on synthetic ISINs (XX00000SYNT3 on Xetra, the LSE and Frankfurt; XX0000FAKEW1 on Warsaw): SupportedExchange accepts exactly WAR, LSE, XETRA and Euronext's PA, AS, BR, LS, IR, MI, OL; Pick returns the one listing on a supported exchange, honours an asked exchange, collapses repeated tickers, and fails with ErrAmbiguous or ErrNotListed otherwise; Ticker joins code and exchange; Quote keeps EUR, PLN and GBP and turns GBX 8412.5 into exactly 84.125 GBP, refusing an empty or non-ISO currency. A false pass would need Pick to guess between exchanges, which the ambiguous case rules out.