---
title: Each sync request to a connector must carry the holdings and open lots of the accounts bound to that instance.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-117-1
  - lane:strict
text_ref: design-decisions:D-117#1
origin:
  kind: import
  ref: design-decisions:D-117#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: Each sync request to a connector must carry the holdings and open lots of the accounts bound to that instance.
logic_claims:
  - CLAIM-1A300E8ADF318010
semantic_clauses:
  - Each sync request to a connector must carry the holdings and open lots of the accounts bound to that instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ca3ca2a0e778b53f7002da7ef0588a6402511101757ed5df4fcdc35cee968bca
semantic_inventory:
  - claim_key: CLAIM-1A300E8ADF318010
    claim_text: Each sync request to a connector must carry the holdings and open lots of the accounts bound to that instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 109
id: REQ-ingest-sync-carries-bound-holdings-lots
type: req
---
Each sync request to a connector must carry the holdings and open lots of the accounts bound to that instance.

## Context

The treasury-bonds connector is stateless and does not know the lots the owner entered, so nothing booked the coupons of COI, ROR and DOR lots (question #19 on docs/implementation-questions.md). Rather than have the core compute coupons, which would put bond-specific logic in the core, the author decided on 2026-10-31 that every sync request carries what the instance's bound accounts hold, as D-93's request carries the instruments to price (REQ-ingest-market-data-sync-instruments). The field is generic: any connector may use it, and the treasury-bonds connector uses it to deliver coupons (REQ-treasurybonds-delivers-coupons).

## Source

> Each sync request to a connector carries the holdings and open lots of the accounts bound to that instance, as D-93's request carries the instruments to price. The treasury-bonds connector uses them to deliver each coupon that fell due since its last sync as interest (D-101), computed with the Polish pack's bond rules, so the core holds no bond logic.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-117#1
