---
title: Ticker confirmations hold per institution and corrections keep versions over PostgreSQL
status: active
sourceFile: internal/ingest/postgres/tickers_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-02T06:59:53.216Z'
id: TEST-ingest-postgres-tickers
type: test
---
Runs the ingest store on a fresh PostgreSQL database: two XTB instances hold a buy and a holding of one ticker; one confirmation through either records it for the institution and returns both instances' waiting rows, which are then marked booked; the pending list drops it, another ISIN is refused, an unknown ticker too; a correction records a new version and returns every row with the previous listing; the confirmed list shows the corrected ISIN with its rows booked; a ticker never confirmed can't be corrected (D-153, D-155).