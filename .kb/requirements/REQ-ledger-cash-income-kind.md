---
title: The ledger must have a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result.
status: open
text_ref: design-decisions:D-152#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-152-1
  - lane:strict
semantic_text: The ledger must have a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result.
logic_claims:
  - CLAIM-B9D395BFE3F32335
semantic_clauses:
  - The ledger must have a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 73f039c32ce83056fcc96317efdc33d271fd0902f989bf14c1846fbbd75eabea
semantic_inventory:
  - claim_key: CLAIM-B9D395BFE3F32335
    claim_text: The ledger must have a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result
    role: normative
    status: modeled
    span:
      start: 0
      end: 148
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-152#1
  recorded_at: '2026-11-01T12:29:07.798Z'
id: REQ-ledger-cash-income-kind
type: req
---
The ledger must have a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result.

## Context

The author recorded it on 2026-11-01. D-101 adds an interest kind for bond coupons (REQ-ledger-interest-kind), and D-24 imports closed CFD results as cash profit or loss (REQ-AUTO-A93F7CCF49981B85). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The ledger has a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result, booked with any tax withheld. Performance counts it as return, not as money paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-152#1
