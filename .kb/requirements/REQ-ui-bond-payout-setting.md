---
title: The built-in UI must offer the payout setting only on a hand-entered bond account's change form and hand the owner's choice to the core, and a form without it must keep the account's setting.
status: closed
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: The built-in UI must offer the payout setting only on a hand-entered bond account's change form and hand the owner's choice to the core, and a form without it must keep the account's setting.
logic_claims:
  - CLAIM-21EDDEAA8EF0F5E4
semantic_clauses:
  - The built-in UI must offer the payout setting only on a hand-entered bond account's change form and hand the owner's choice to the core, and a form without it must keep the account's setting.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b566b282c4971e8f7bd8d803e9d9c79b54ad1041fbebc686cdef17b8349d28e7
semantic_inventory:
  - claim_key: CLAIM-21EDDEAA8EF0F5E4
    claim_text: The built-in UI must offer the payout setting only on a hand-entered bond account's change form and hand the owner's choice to the core, and a form without it must keep the account's setting
    role: normative
    status: modeled
    span:
      start: 0
      end: 190
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:48.645Z'
id: REQ-ui-bond-payout-setting
type: req
---
The built-in UI must offer the payout setting only on a hand-entered bond account's change form and hand the owner's choice to the core, and a form without it must keep the account's setting.

## Context

D-135's setting is the owner's; the UI follows the design system's form rules. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135
