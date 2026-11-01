---
title: Performance must count cash income as return, not as money paid in.
status: open
text_ref: design-decisions:D-152#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-152-3
  - lane:strict
semantic_text: Performance must count cash income as return, not as money paid in.
logic_claims:
  - CLAIM-2BEECA0410E8F560
semantic_clauses:
  - Performance must count cash income as return, not as money paid in.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cbadc1198f1adcac5c80e25867c7e646d78c97e1b3f4f4bd77e29a1b764aa3d3
semantic_inventory:
  - claim_key: CLAIM-2BEECA0410E8F560
    claim_text: Performance must count cash income as return, not as money paid in
    role: normative
    status: modeled
    span:
      start: 0
      end: 66
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-152#3
  recorded_at: '2026-11-01T12:29:32.580Z'
id: REQ-portfolio-cash-income-counts-as-return
type: req
---
Performance must count cash income as return, not as money paid in.

## Context

The author recorded it on 2026-11-01. Performance is the time-weighted return (D-61), in which deposits and withdrawals don't count (REQ-portfolio-twr-comparison). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The ledger has a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result, booked with any tax withheld. Performance counts it as return, not as money paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-152#3
