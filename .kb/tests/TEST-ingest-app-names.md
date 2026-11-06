---
title: A sync's names and a confirmed ticker's name reach the ledger
status: active
sourceFile: internal/ingest/app/names_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:58.121Z'
id: TEST-ingest-app-names
type: test
---
Connectors.Sync hands a delivery's names to Names as the instance's; Tickers.Confirm names the confirmed ISIN by the institution's ticker name, as its instance's, with a fake store and namer.