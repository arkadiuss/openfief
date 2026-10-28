---
title: In M0, fees must be posted to the fee ledger account and must not be added to lot cost.
status: open
text_ref: design-decisions:D-34#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-34-3
  - lane:strict
semantic_text: In M0, fees must be posted to the fee ledger account and must not be added to lot cost.
logic_claims:
  - CLAIM-C1BFDFC332356BB8
semantic_clauses:
  - In M0, fees must be posted to the fee ledger account and must not be added to lot cost.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1d2763f733db1539a0443ca64eb50b6f4b6f5c46b13a335f515398aa55bfb561
semantic_inventory:
  - claim_key: CLAIM-C1BFDFC332356BB8
    claim_text: In M0, fees must be posted to the fee ledger account and must not be added to lot cost
    role: normative
    status: modeled
    span:
      start: 0
      end: 86
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-34#3
  recorded_at: '2026-10-28T06:28:32.560Z'
id: REQ-decision-d34-3-fees-not-in-lot-cost
type: req
---
In M0, fees must be posted to the fee ledger account and must not be added to lot cost.

## Context

The author decided D-34 in docs/DESIGN.md section 3.3; this requirement imports its claim D-34#3 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> In M0 fees go to the fee ledger account and are not added to lot cost.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-34#3
