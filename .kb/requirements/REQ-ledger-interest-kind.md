---
title: The ledger must have an interest transaction kind.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-101-1
  - lane:strict
text_ref: design-decisions:D-101#1
origin:
  kind: import
  ref: design-decisions:D-101#1
  recorded_at: '2026-10-29T09:57:15.954Z'
semantic_text: The ledger must have an interest transaction kind.
logic_claims:
  - CLAIM-48EE28A1A07C1B91
semantic_clauses:
  - The ledger must have an interest transaction kind.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: b942adb4d6bc8e44b78391840c60bae90bc6a462be3dcc9c019d1c7a47269fed
semantic_inventory:
  - claim_key: CLAIM-48EE28A1A07C1B91
    claim_text: The ledger must have an interest transaction kind
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
id: REQ-ledger-interest-kind
type: req
---
The ledger must have an interest transaction kind.

## Context

D-19 books every transaction as double-entry postings, and REQ-ledger-event-postings lists the kinds booked so far: buys, sells, dividends with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots. D-101 adds an interest kind, which bond coupons use (REQ-ledger-coupons-as-interest). The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> The ledger has an interest transaction kind. Bond coupons, such as COI's yearly interest, are booked as interest, not as dividends.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-101#1
