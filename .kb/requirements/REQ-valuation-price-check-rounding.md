---
title: A price check must give the cross-check price's relative difference from the price rounded half-even to six decimal places, and must decide whether it is flagged on the exact difference, a difference equal to the threshold being matched.
status: open
text_ref: docs/DESIGN.md#D-110
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-110
  - lane:strict
semantic_text: A price check must give the cross-check price's relative difference from the price rounded half-even to six decimal places, and must decide whether it is flagged on the exact difference, a difference equal to the threshold being matched.
logic_claims:
  - CLAIM-41FEC91D392F6550
semantic_clauses:
  - A price check must give the cross-check price's relative difference from the price rounded half-even to six decimal places, and must decide whether it is flagged on the exact difference, a difference equal to the threshold being matched.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 843f9f96f9a41d915ea4bbb5cffc8f5387778b8adc03444ca3331d3aac584d52
semantic_inventory:
  - claim_key: CLAIM-41FEC91D392F6550
    claim_text: A price check must give the cross-check price's relative difference from the price rounded half-even to six decimal places, and must decide whether it is flagged on the exact difference, a difference equal to the threshold being matched
    role: normative
    status: modeled
    span:
      start: 0
      end: 236
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:41.322Z'
id: REQ-valuation-price-check-rounding
type: req
---
A price check must give the cross-check price's relative difference from the price rounded half-even to six decimal places, and must decide whether it is flagged on the exact difference, a difference equal to the threshold being matched.

## Context

D-110 fixes when a check is flagged but not how its difference is shown. This restores, as a refinement, the detail of the superseded REQ-valuation-price-check: rounding is for display only. Reason for six places not stated by the author; it is the implementation's choice.

## Source

> A cross-check price is flagged when it differs from the stored price by more than 1% of it, in either direction. The owner can set a different threshold for a single instrument, such as a thinly traded Warsaw listing.

docs/DESIGN.md#D-110
