---
title: The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid.
logic_claims:
  - CLAIM-FC88B8A6209245E9
semantic_clauses:
  - The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 02d1f90a653be50c6d54f9c01ead4a88f7431ea678014635e4055d98fe09bebb
semantic_inventory:
  - claim_key: CLAIM-FC88B8A6209245E9
    claim_text: The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid
    role: normative
    status: modeled
    span:
      start: 0
      end: 363
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:19.768Z'
id: REQ-ingest-connector-batches
type: req
---
The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid.

## Context

D-49: the core calls connectors to sync and to import files, and owns the schedules, retries and import batches; uploads go through the core to the right connector. Re-imports are idempotent (D-18), so a retry with the same batch ID is safe. The schedule and the stored watermark are the caller's (a later task). Refines REQ-ingest-core-calls-connectors, REQ-ingest-core-owns-schedules, REQ-ingest-uploads-through-core and REQ-AUTO-41A1F4D687DAA771. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. [...] File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md D-49
