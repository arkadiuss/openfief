---
title: Uploads choose the start and hand the owner's opening cash on; confirmed opening lots join the seeding
status: active
sourceFile: internal/ingest/app/start_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:18.398Z'
id: TEST-ingest-app-start
type: test
---
Runs app.Uploads and app.Tickers over in-memory fakes of the connector, ledger, history and ticker store: an upload for a new account hands the opening cash to the connector in opening mode, an account started from its whole history takes transactions and refuses an opening override, opening cash with a transactions upload or for a seeded account is refused before the connector is called, and an unreadable history fails the upload. Confirming a ticker whose waiting entries include an opening lot adds them through AddOpeningLots, while a buy alone is imported.