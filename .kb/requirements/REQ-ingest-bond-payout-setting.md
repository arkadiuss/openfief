---
title: The owner must be able to set a hand-entered bond account's payout setting when creating or changing the account through AccountService, in force from that day, and a paid-out setting on an account at any other institution must be refused as invalid.
status: closed
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: The owner must be able to set a hand-entered bond account's payout setting when creating or changing the account through AccountService, in force from that day, and a paid-out setting on an account at any other institution must be refused as invalid.
logic_claims:
  - CLAIM-2856D7445FEE913C
semantic_clauses:
  - The owner must be able to set a hand-entered bond account's payout setting when creating or changing the account through AccountService, in force from that day, and a paid-out setting on an account at any other institution must be refused as invalid.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7dd8d9ba69998d68faae0416651af6b254119b5386a7d23078ef53ed3b384251
semantic_inventory:
  - claim_key: CLAIM-2856D7445FEE913C
    claim_text: The owner must be able to set a hand-entered bond account's payout setting when creating or changing the account through AccountService, in force from that day, and a paid-out setting on an account at any other institution must be refused as invalid
    role: normative
    status: modeled
    span:
      start: 0
      end: 249
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:09.281Z'
id: REQ-ingest-bond-payout-setting
type: req
---
The owner must be able to set a hand-entered bond account's payout setting when creating or changing the account through AccountService, in force from that day, and a paid-out setting on an account at any other institution must be refused as invalid.

## Context

D-135: the setting matches the option the bond registry offers, so only an account at the treasury-bonds institution may pay out. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
