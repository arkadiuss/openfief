---
title: The built-in UI must offer, on a hand-entered bond account's change form, the payout setting and the day the change takes effect from, and hand both to the core.
status: open
text_ref: docs/DESIGN.md#D-162
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-162
  - lane:strict
semantic_text: The built-in UI must offer, on a hand-entered bond account's change form, the payout setting and the day the change takes effect from, and hand both to the core.
logic_claims:
  - CLAIM-13D99D298B0EC36E
semantic_clauses:
  - The built-in UI must offer, on a hand-entered bond account's change form, the payout setting and the day the change takes effect from, and hand both to the core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4b96c273a32affeb4d33bef0676e083a12075e2deb1826a1429c3f93bee76004
semantic_inventory:
  - claim_key: CLAIM-13D99D298B0EC36E
    claim_text: The built-in UI must offer, on a hand-entered bond account's change form, the payout setting and the day the change takes effect from, and hand both to the core
    role: normative
    status: modeled
    span:
      start: 0
      end: 160
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:24.976Z'
id: REQ-ui-bond-payout-change-date
type: req
---
The built-in UI must offer, on a hand-entered bond account's change form, the payout setting and the day the change takes effect from, and hand both to the core.

## Context

D-162 has the owner give the day a payout change takes effect from; the work package asked for a date field in the UI. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Why this changed

The author decided this in D-162, which replaces the narrower reading REQ-ui-bond-payout-setting recorded before the decision existed, so that requirement is closed.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

docs/DESIGN.md#D-162
