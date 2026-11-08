---
title: The core re-keys a provisional instrument everywhere, the ledger last
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:28.718Z'
id: SCEN-ingest-rekey-provisional-instrument
type: scenario
---
Given a sync delivers the registry ISIN of an instrument held under its provisional ISIN, when the core imports the delivery, then after the ledger import it moves what every context stores under the instrument to the registry's, the ledger last, and the syncs after name it no more and change nothing; an invalid one refuses the delivery and an undeclared one is dropped.