---
title: The ledger must withdraw the holding assertions a ticker's ISIN correction replaces, removing their checks and never a cash assertion or a transaction.
status: open
text_ref: docs/DESIGN.md#D-153
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-153
  - lane:strict
semantic_text: The ledger must withdraw the holding assertions a ticker's ISIN correction replaces, removing their checks and never a cash assertion or a transaction.
logic_claims:
  - CLAIM-A2D9290BAAED8441
semantic_clauses:
  - The ledger must withdraw the holding assertions a ticker's ISIN correction replaces, removing their checks and never a cash assertion or a transaction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 89af6798fe7e234c28c3e8e4718dd9a969764f612a299d9949cea235572b3c73
semantic_inventory:
  - claim_key: CLAIM-A2D9290BAAED8441
    claim_text: The ledger must withdraw the holding assertions a ticker's ISIN correction replaces, removing their checks and never a cash assertion or a transaction
    role: normative
    status: modeled
    span:
      start: 0
      end: 150
origin:
  kind: agent
  recorded_at: '2026-11-02T07:01:08.378Z'
id: REQ-ledger-withdraw-holding-assertions
type: req
---
The ledger must withdraw the holding assertions a ticker's ISIN correction replaces, removing their checks and never a cash assertion or a transaction.

## Context

D-153 supersedes the transactions a wrong ISIN booked; the holdings asserted of the wrong instrument no longer hold, so their checks are withdrawn rather than left flagged. This is this implementation's reading.

## Source

> **[D-153]** A ticker's confirmed ISIN (D-116) holds for every account at the same institution. The owner can correct a wrong one, which supersedes the transactions it booked (D-18).

docs/DESIGN.md#D-153
