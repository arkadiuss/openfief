---
title: The owner must be able to change a hand-entered bond account's payout setting from a day given, today when none is and never later, and the payout withdrawals booked from that day that the new setting no longer makes must be superseded.
status: closed
text_ref: docs/DESIGN.md#D-162
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-162
  - lane:strict
semantic_text: The owner must be able to change a hand-entered bond account's payout setting from a day given, today when none is and never later, and the payout withdrawals booked from that day that the new setting no longer makes must be superseded.
logic_claims:
  - CLAIM-48EFB5218155823B
semantic_clauses:
  - The owner must be able to change a hand-entered bond account's payout setting from a day given, today when none is and never later, and the payout withdrawals booked from that day that the new setting no longer makes must be superseded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b468dd53f5276e1189ea09cb19cce3ae4b4888c101bf3f4d2cfc19d9b5bf60f7
semantic_inventory:
  - claim_key: CLAIM-48EFB5218155823B
    claim_text: The owner must be able to change a hand-entered bond account's payout setting from a day given, today when none is and never later, and the payout withdrawals booked from that day that the new setting no longer makes must be superseded
    role: normative
    status: modeled
    span:
      start: 0
      end: 235
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:03.061Z'
id: REQ-ingest-bond-payout-change-from-day
type: req
---
The owner must be able to change a hand-entered bond account's payout setting from a day given, today when none is and never later, and the payout withdrawals booked from that day that the new setting no longer makes must be superseded.

## Context

D-162 has a change take effect from a day the owner gives and the coupons and maturities after it booked anew, superseding the earlier ones. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Why this changed

The author decided this in D-162, which replaces the narrower reading REQ-ingest-bond-payout-setting recorded before the decision existed, so that requirement is closed.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

docs/DESIGN.md#D-162
