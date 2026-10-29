---
title: The core must serve BondLotService to view plugins only, read a lot's rates from percent into fractions, and answer each failure with its own Connect code, never exposing an internal cause.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must serve BondLotService to view plugins only, read a lot's rates from percent into fractions, and answer each failure with its own Connect code, never exposing an internal cause.
logic_claims:
  - CLAIM-CF275AB923D0045C
semantic_clauses:
  - The core must serve BondLotService to view plugins only, read a lot's rates from percent into fractions, and answer each failure with its own Connect code, never exposing an internal cause.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1fb7d47f4ce2af2915e962017437d543dc30edcca947107013bc4edce2ad387c
semantic_inventory:
  - claim_key: CLAIM-CF275AB923D0045C
    claim_text: The core must serve BondLotService to view plugins only, read a lot's rates from percent into fractions, and answer each failure with its own Connect code, never exposing an internal cause
    role: normative
    status: modeled
    span:
      start: 0
      end: 188
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:41.777Z'
id: REQ-ingest-bond-lot-service
type: req
---
The core must serve BondLotService to view plugins only, read a lot's rates from percent into fractions, and answer each failure with its own Connect code, never exposing an internal cause.

## Context

D-49 routes manual entries from the UI through the core; the service shape follows the ModelPortfolioService (view plugins only, owner session in the request), which still awaits the author's confirmation. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
