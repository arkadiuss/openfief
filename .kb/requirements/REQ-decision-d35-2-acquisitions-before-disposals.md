---
title: Within one date, acquisitions must apply before disposals.
status: open
text_ref: design-decisions:D-35#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-35-2
  - lane:strict
semantic_text: Within one date, acquisitions must apply before disposals.
logic_claims:
  - CLAIM-DBD19DF545D96C12
semantic_clauses:
  - Within one date, acquisitions must apply before disposals.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 11333baf387f4f5a705439137b1433fa1036b1b4efdf6c91543b30c7a2c1b5d7
semantic_inventory:
  - claim_key: CLAIM-DBD19DF545D96C12
    claim_text: Within one date, acquisitions must apply before disposals
    role: normative
    status: modeled
    span:
      start: 0
      end: 57
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-35#2
  recorded_at: '2026-10-28T06:28:47.776Z'
id: REQ-decision-d35-2-acquisitions-before-disposals
type: req
---
Within one date, acquisitions must apply before disposals.

## Context

The author decided D-35 in docs/DESIGN.md section 3.3; this requirement imports its claim D-35#2 as a firm rule, so changing it needs the author. The decision gives the reason: broker rows carry no reliable time of day, so the order of a day's rows cannot be trusted. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Within one date, acquisitions apply before disposals, because broker rows carry no reliable time of day.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-35#2
