---
title: The owner must be able to correct a ticker's wrong confirmed ISIN, which supersedes the transactions it booked.
status: open
text_ref: design-decisions:D-153#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-153-2
  - lane:strict
semantic_text: The owner must be able to correct a ticker's wrong confirmed ISIN, which supersedes the transactions it booked.
logic_claims:
  - CLAIM-C988B01E8BE311E8
semantic_clauses:
  - The owner must be able to correct a ticker's wrong confirmed ISIN, which supersedes the transactions it booked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a07cec4e483dc1091b6acb9d74659f526efe27b805fe565a90b2ab01560213f3
semantic_inventory:
  - claim_key: CLAIM-C988B01E8BE311E8
    claim_text: The owner must be able to correct a ticker's wrong confirmed ISIN, which supersedes the transactions it booked
    role: normative
    status: modeled
    span:
      start: 0
      end: 110
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-153#2
  recorded_at: '2026-11-01T12:29:50.072Z'
id: REQ-ingest-confirmed-isin-owner-corrects
type: req
---
The owner must be able to correct a ticker's wrong confirmed ISIN, which supersedes the transactions it booked.

## Context

The author recorded it on 2026-11-01. D-18 keeps the transaction log append-only and versioned, so a correction supersedes the transactions booked with the wrong ISIN instead of erasing them. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A ticker's confirmed ISIN (D-116) holds for every account at the same institution. The owner can correct a wrong one, which supersedes the transactions it booked (D-18).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-153#2
