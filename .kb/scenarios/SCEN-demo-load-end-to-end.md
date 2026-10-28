---
title: The loader fills a demo database through the real pipelines
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:45.270Z'
id: SCEN-demo-load-end-to-end
type: scenario
---
Given an empty PostgreSQL database, when the loader runs for 2026-10-28, then the IBKR account is seeded through the ibkrflex connector and the ingest service and gets its monthly batches, XTB and the bond account are seeded and imported through the ledger, prices, FX rates, CPI readings, bond terms and the IKE wrapper are recorded, and the River workers value every account through 2026-10-28 with no input missing before the loader exits.