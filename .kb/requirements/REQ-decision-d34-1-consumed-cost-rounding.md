---
title: Selling part of a lot must round the consumed cost to two decimal places, half-even.
status: open
text_ref: design-decisions:D-34#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-34-1
  - lane:strict
semantic_text: Selling part of a lot must round the consumed cost to two decimal places, half-even.
logic_claims:
  - CLAIM-F1AFE51C08440220
semantic_clauses:
  - Selling part of a lot must round the consumed cost to two decimal places, half-even.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bc452ac506aa4d6e751814a5563c541ca2278a60fe99775c8ec7fefc4663be60
semantic_inventory:
  - claim_key: CLAIM-F1AFE51C08440220
    claim_text: Selling part of a lot must round the consumed cost to two decimal places, half-even
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-34#1
  recorded_at: '2026-10-28T06:28:16.326Z'
id: REQ-decision-d34-1-consumed-cost-rounding
type: req
---
Selling part of a lot must round the consumed cost to two decimal places, half-even.

## Context

The author decided D-34 in docs/DESIGN.md section 3.3; this requirement imports its claim D-34#1 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Selling part of a lot rounds the consumed cost to two decimal places, half-even, and the lot keeps the exact remainder, so consumed plus remaining cost always equals the original cost.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-34#1
