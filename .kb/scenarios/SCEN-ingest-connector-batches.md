---
title: The core imports what a connector instance delivers as one import batch
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:44.500Z'
id: SCEN-ingest-connector-batches
type: scenario
---
Given connector instance xtb-main and batch ID b-1, when the core syncs it with a watermark or hands it an uploaded file, then the core imports the delivery's transactions and balance assertions into the ledger and stores its prices, FX rates and index readings with xtb-main as source, as one import batch with source sync or file, and returns the next watermark of a sync; when the connector fails or the delivery is invalid, nothing is imported, and the caller retries with the same batch ID.