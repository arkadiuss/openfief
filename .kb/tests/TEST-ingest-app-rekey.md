---
title: A sync's registry ISINs re-key each provisional instrument with every context, the ledger last
status: active
sourceFile: internal/ingest/app/rekey_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:14.316Z'
id: TEST-ingest-app-rekey
type: test
---
Drives app.Connectors.Sync with a fake caller whose delivery carries registry ISINs and two recording rekeyers: each instrument is re-keyed by every rekeyer in order after the ledger import, undeclared ones are dropped with the rest imported, an invalid one refuses the delivery before any import, a failing re-key fails the batch, and no rekeyer is ErrNoRekeyer. HoldingsScope.ProvisionalInstruments names each provisional instrument of the asked accounts once and the scheduler sends them in the sync call. A false pass would re-key the ledger before the other stores.