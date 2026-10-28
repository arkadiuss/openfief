---
title: The core service must migrate the ledger, valuation, River and identity schemas, run the River revaluation workers, and serve the read API as Connect and REST under /v1 with script tokens as the accepted bearer tokens.
status: open
text_ref: docs/DESIGN.md#D-2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-2
  - lane:strict
semantic_text: The core service must migrate the ledger, valuation, River and identity schemas, run the River revaluation workers, and serve the read API as Connect and REST under /v1 with script tokens as the accepted bearer tokens.
logic_claims:
  - CLAIM-E796888395449432
semantic_clauses:
  - The core service must migrate the ledger, valuation, River and identity schemas, run the River revaluation workers, and serve the read API as Connect and REST under /v1 with script tokens as the accepted bearer tokens.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7a9e8b1a2f5abffa7852915c13dba27c7dceed95eae1480d540bd8b9f9b1737c
semantic_inventory:
  - claim_key: CLAIM-E796888395449432
    claim_text: The core service must migrate the ledger, valuation, River and identity schemas, run the River revaluation workers, and serve the read API as Connect and REST under /v1 with script tokens as the accepted bearer tokens
    role: normative
    status: modeled
    span:
      start: 0
      end: 217
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:03.811Z'
id: REQ-core-serves-read-api
type: req
---
The core service must migrate the ledger, valuation, River and identity schemas, run the River revaluation workers, and serve the read API as Connect and REST under /v1 with script tokens as the accepted bearer tokens.

## Context

D-2 makes the core the system of record; D-36 runs revaluations on River; D-50 serves the read API as Connect and REST; D-27 gives scripts tokens. The author asked to see the core running on the demo. Connector calls (D-49) wait for the first connector plugin. Modeled by Claude Code for the core task: the author asked to see OpenFief's core running on the demo; the author has not reviewed this requirement yet.

## Source

> OpenFief's own core is the system of record for transactions, tax lots and valuations.

docs/DESIGN.md D-2
