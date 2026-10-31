---
title: The core syncs connector instances and imports files through them as import batches
status: active
sourceFile: internal/ingest/app/connectors_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:24.145Z'
id: TEST-ingest-app-connectors
type: test
verification_scope: unit
---
Runs app.Connectors over in-memory fakes of the connector caller, the ledger importer and the price, FX rate and index reading stores, on synthetic data. A sync of instance xtb-main as batch b-1 asks the connector with the given mode and watermark, imports the transactions with keys scoped to plugin:xtb-main, stores the price, FX rate and index reading with xtb-main as source whatever the connector set, and returns the completed batch (source sync) with the next watermark. An imported file reaches the connector with its name and contents and becomes a batch with source file; a delivery without valuation inputs calls no store. A failing connector, a delivery without institution, a missing mode or an invalid index reading imports nothing; a ledger or store failure fails the batch.