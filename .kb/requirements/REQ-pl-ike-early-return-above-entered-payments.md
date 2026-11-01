---
title: An IKE whose earlier history is unknown must have its early return taxed above the total payments in the owner entered.
status: open
text_ref: design-decisions:D-149#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-149-2
  - lane:strict
semantic_text: An IKE whose earlier history is unknown must have its early return taxed above the total payments in the owner entered.
logic_claims:
  - CLAIM-50E06FAC00A1C32D
semantic_clauses:
  - An IKE whose earlier history is unknown must have its early return taxed above the total payments in the owner entered.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f8421d24adebc9717d64dcb5155c4554aadc2d25c1b89d8fb401ae50ca5cd985
semantic_inventory:
  - claim_key: CLAIM-50E06FAC00A1C32D
    claim_text: An IKE whose earlier history is unknown must have its early return taxed above the total payments in the owner entered
    role: normative
    status: modeled
    span:
      start: 0
      end: 118
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-149#2
  recorded_at: '2026-11-01T12:27:58.807Z'
id: REQ-pl-ike-early-return-above-entered-payments
type: req
---
An IKE whose earlier history is unknown must have its early return taxed above the total payments in the owner entered.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #44. D-128 taxes an IKE's early return on what is withdrawn above what was paid in (REQ-pl-ike-early-return-above-payments). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> For an IKE whose earlier history is unknown (D-23), the owner enters its total payments in, which its early return is taxed above.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-149#2
