---
title: An instrument's page, its next page of transactions, the owner's rename and the price check forms returning to it
status: active
sourceFile: internal/views/ui/instrument_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T10:42:50.998Z'
id: TEST-ui-instrument
type: test
---
Serves the UI against in-memory Connect fakes of the read API and the InstrumentService: the page shows the display name, tickers and ISIN, the breadcrumb through the account it was opened from, the latest price, a price line per source with the primary marked, holdings, transactions with their accounts and a next page, price checks with forms that return to the page, the OKI class and the open items; a rename reaches the core trimmed, a refusal shows its reason, empty, overlong and malformed names never reach the core, a read session is refused.