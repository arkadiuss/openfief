---
title: The treasury-bonds connector must deliver each coupon that fell due since its last sync as interest, computed with the Polish pack's bond rules from the holdings and open lots its sync request carries, so that the core holds no bond logic.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-117-2
  - lane:strict
text_ref: design-decisions:D-117#2
origin:
  kind: import
  ref: design-decisions:D-117#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The treasury-bonds connector must deliver each coupon that fell due since its last sync as interest, computed with the Polish pack's bond rules from the holdings and open lots its sync request carries, so that the core holds no bond logic.
logic_claims:
  - CLAIM-CDEAF7E0F8304C2C
semantic_clauses:
  - The treasury-bonds connector must deliver each coupon that fell due since its last sync as interest, computed with the Polish pack's bond rules from the holdings and open lots its sync request carries, so that the core holds no bond logic.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 78136912e5456fcad4aa3b9eea1b18dcd8a05f25b624d629defbd0acb5710aa0
semantic_inventory:
  - claim_key: CLAIM-CDEAF7E0F8304C2C
    claim_text: The treasury-bonds connector must deliver each coupon that fell due since its last sync as interest, computed with the Polish pack's bond rules from the holdings and open lots its sync request carries, so that the core holds no bond logic
    role: normative
    status: modeled
    span:
      start: 0
      end: 238
id: REQ-treasurybonds-delivers-coupons
type: req
---
The treasury-bonds connector must deliver each coupon that fell due since its last sync as interest, computed with the Polish pack's bond rules from the holdings and open lots its sync request carries, so that the core holds no bond logic.

## Context

D-101 books bond coupons, such as COI's yearly interest, as interest (REQ-ledger-coupons-as-interest), and the pack's coupon types count only the current period's interest in a lot's value, so a paid coupon must reach the account's cash or the account loses it. Nothing booked the coupons of lots the owner entered, because the connector is stateless and didn't know the lots (question #19 on docs/implementation-questions.md). The options were the core computing each due coupon or the owner entering coupons by hand; on 2026-10-31 the author chose neither and decided D-117, replacing the core-computes option because they didn't want bond-specific logic in the core. The connector learns the lots from its sync request (REQ-ingest-sync-carries-bound-holdings-lots).

## Source

> Each sync request to a connector carries the holdings and open lots of the accounts bound to that instance, as D-93's request carries the instruments to price. The treasury-bonds connector uses them to deliver each coupon that fell due since its last sync as interest (D-101), computed with the Polish pack's bond rules, so the core holds no bond logic.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-117#2
