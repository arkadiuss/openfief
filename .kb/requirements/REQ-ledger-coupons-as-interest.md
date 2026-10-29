---
title: Bond coupons, such as COI's yearly interest, must be booked as interest, not as dividends.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-101-2
  - lane:strict
text_ref: design-decisions:D-101#2
origin:
  kind: import
  ref: design-decisions:D-101#2
  recorded_at: '2026-10-29T09:57:58.835Z'
semantic_text: Bond coupons, such as COI's yearly interest, must be booked as interest, not as dividends.
logic_claims:
  - CLAIM-9A19B776332F14F0
semantic_clauses:
  - Bond coupons, such as COI's yearly interest, must be booked as interest, not as dividends.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0a81888f36d2125219e682c284f3fc9eefa53cd574ff9b990753f5e3143e7fbc
semantic_inventory:
  - claim_key: CLAIM-9A19B776332F14F0
    claim_text: Bond coupons, such as COI's yearly interest, must be booked as interest, not as dividends
    role: normative
    status: modeled
    span:
      start: 0
      end: 89
id: REQ-ledger-coupons-as-interest
type: req
---
Bond coupons, such as COI's yearly interest, must be booked as interest, not as dividends.

## Context

D-101 gives the ledger an interest transaction kind (REQ-ledger-interest-kind) and books bond coupons with it, such as the yearly interest a COI bond pays, rather than as dividends. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> The ledger has an interest transaction kind. Bond coupons, such as COI's yearly interest, are booked as interest, not as dividends.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-101#2
