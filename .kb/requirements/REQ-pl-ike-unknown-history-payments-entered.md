---
title: For an IKE whose earlier history is unknown, the owner must enter its total payments in.
status: open
text_ref: design-decisions:D-149#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-149-1
  - lane:strict
semantic_text: For an IKE whose earlier history is unknown, the owner must enter its total payments in.
logic_claims:
  - CLAIM-F69470FDD726029B
semantic_clauses:
  - For an IKE whose earlier history is unknown, the owner must enter its total payments in.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d65cd463cbd937b604fa1d189a3f6ad443203b6f4484c29bda3b454d98765d15
semantic_inventory:
  - claim_key: CLAIM-F69470FDD726029B
    claim_text: For an IKE whose earlier history is unknown, the owner must enter its total payments in
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-149#1
  recorded_at: '2026-11-01T12:27:51.528Z'
id: REQ-pl-ike-unknown-history-payments-entered
type: req
---
For an IKE whose earlier history is unknown, the owner must enter its total payments in.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #44. D-23 seeds accounts from current positions, so an IKE's payments in before go-live may be unknown. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> For an IKE whose earlier history is unknown (D-23), the owner enters its total payments in, which its early return is taxed above.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-149#1
