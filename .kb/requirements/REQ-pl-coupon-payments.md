---
title: The Polish pack must date each coupon of a coupon series (ROR, DOR, COI) on the first day of the next interest period, the last period's excepted, and compute one bond's coupon as the nominal times the period's rate times its months over twelve, rounded half up to grosze.
status: open
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
semantic_text: The Polish pack must date each coupon of a coupon series (ROR, DOR, COI) on the first day of the next interest period, the last period's excepted, and compute one bond's coupon as the nominal times the period's rate times its months over twelve, rounded half up to grosze.
logic_claims:
  - CLAIM-96DB6C506925A66D
semantic_clauses:
  - The Polish pack must date each coupon of a coupon series (ROR, DOR, COI) on the first day of the next interest period, the last period's excepted, and compute one bond's coupon as the nominal times the period's rate times its months over twelve, rounded half up to grosze.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fe21af5c50bb0910c25c4a866210996afc31fc66db5f9d988b4283ed6be74ec3
semantic_inventory:
  - claim_key: CLAIM-96DB6C506925A66D
    claim_text: The Polish pack must date each coupon of a coupon series (ROR, DOR, COI) on the first day of the next interest period, the last period's excepted, and compute one bond's coupon as the nominal times the period's rate times its months over twelve, rounded half up to grosze
    role: normative
    status: modeled
    span:
      start: 0
      end: 271
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:30.162Z'
id: REQ-pl-coupon-payments
type: req
---
The Polish pack must date each coupon of a coupon series (ROR, DOR, COI) on the first day of the next interest period, the last period's excepted, and compute one bond's coupon as the nominal times the period's rate times its months over twelve, rounded half up to grosze.

## Context

D-117 has the treasury-bonds connector compute each coupon with the Polish pack's bond rules, so the rules live in internal/jurisdiction/pl, next to ValueLot, which counts a whole period the same way. The last period's interest is paid with the redemption at maturity, which OpenFief does not book yet (implementation questions). Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Each sync request to a connector carries the holdings and open lots of the accounts bound to that instance, as D-93's request carries the instruments to price. The treasury-bonds connector uses them to deliver each coupon that fell due since its last sync as interest (D-101), computed with the Polish pack's bond rules, so the core holds no bond logic.

docs/DESIGN.md#D-117
