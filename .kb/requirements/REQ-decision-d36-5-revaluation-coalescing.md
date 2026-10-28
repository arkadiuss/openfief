---
title: Pending revaluations must coalesce per account from the earliest affected date.
status: open
text_ref: design-decisions:D-36#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-5
  - lane:strict
semantic_text: Pending revaluations must coalesce per account from the earliest affected date.
logic_claims:
  - CLAIM-6EDC5169CEC95EED
semantic_clauses:
  - Pending revaluations must coalesce per account from the earliest affected date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1851a1884c38c2cdbaf0f1e38fcd7320cd038903807ae9a63eb7af181fadae96
semantic_inventory:
  - claim_key: CLAIM-6EDC5169CEC95EED
    claim_text: Pending revaluations must coalesce per account from the earliest affected date
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#5
  recorded_at: '2026-10-28T06:29:43.898Z'
id: REQ-decision-d36-5-revaluation-coalescing
type: req
---
Pending revaluations must coalesce per account from the earliest affected date.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#5 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Pending revaluations coalesce per account from the earliest affected date.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#5
