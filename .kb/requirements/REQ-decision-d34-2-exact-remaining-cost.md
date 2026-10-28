---
title: After a partial sale the lot must keep the exact remaining cost, so consumed plus remaining cost always equals the original cost.
status: open
text_ref: design-decisions:D-34#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-34-2
  - lane:strict
semantic_text: After a partial sale the lot must keep the exact remaining cost, so consumed plus remaining cost always equals the original cost.
logic_claims:
  - CLAIM-F77FB37CF9B942BB
semantic_clauses:
  - After a partial sale the lot must keep the exact remaining cost, so consumed plus remaining cost always equals the original cost.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1ed0a24d4a775334bdc720b72b60e703671d63bd405a502924476651140d3a12
semantic_inventory:
  - claim_key: CLAIM-F77FB37CF9B942BB
    claim_text: After a partial sale the lot must keep the exact remaining cost, so consumed plus remaining cost always equals the original cost
    role: normative
    status: modeled
    span:
      start: 0
      end: 128
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-34#2
  recorded_at: '2026-10-28T06:28:24.722Z'
id: REQ-decision-d34-2-exact-remaining-cost
type: req
---
After a partial sale the lot must keep the exact remaining cost, so consumed plus remaining cost always equals the original cost.

## Context

The author decided D-34 in docs/DESIGN.md section 3.3; this requirement imports its claim D-34#2 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Selling part of a lot rounds the consumed cost to two decimal places, half-even, and the lot keeps the exact remainder, so consumed plus remaining cost always equals the original cost.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-34#2
