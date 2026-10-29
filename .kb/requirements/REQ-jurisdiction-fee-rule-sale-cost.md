---
title: A sale's cost under a fee rule must be the rate of the value sold but at least the minimum, plus the conversion rate of the value for an instrument in a foreign currency, rounded half up to the minor unit, nothing on the zero rule, with rates from 0 up to 1 and a minimum that is not negative.
status: open
text_ref: docs/DESIGN.md#D-78
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-78
  - lane:strict
semantic_text: A sale's cost under a fee rule must be the rate of the value sold but at least the minimum, plus the conversion rate of the value for an instrument in a foreign currency, rounded half up to the minor unit, nothing on the zero rule, with rates from 0 up to 1 and a minimum that is not negative.
logic_claims:
  - CLAIM-638BA49E9F4577C4
semantic_clauses:
  - A sale's cost under a fee rule must be the rate of the value sold but at least the minimum, plus the conversion rate of the value for an instrument in a foreign currency, rounded half up to the minor unit, nothing on the zero rule, with rates from 0 up to 1 and a minimum that is not negative.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f7203781428cfee81c4cee57f3a245657593ff44c1e1aef4135e926ff152853e
semantic_inventory:
  - claim_key: CLAIM-638BA49E9F4577C4
    claim_text: A sale's cost under a fee rule must be the rate of the value sold but at least the minimum, plus the conversion rate of the value for an instrument in a foreign currency, rounded half up to the minor unit, nothing on the zero rule, with rates from 0 up to 1 and a minimum that is not negative
    role: normative
    status: modeled
    span:
      start: 0
      end: 292
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:37.290Z'
id: REQ-jurisdiction-fee-rule-sale-cost
type: req
---
A sale's cost under a fee rule must be the rate of the value sold but at least the minimum, plus the conversion rate of the value for an instrument in a foreign currency, rounded half up to the minor unit, nothing on the zero rule, with rates from 0 up to 1 and a minimum that is not negative.

## Context

D-78 fixes the rule's shape. Applying the minimum to the commission only, converting the minimum at the day's rate and rounding half up are this package's choices. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-78]** Liquidation value estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency. The rule is zero until set. Withdrawing from IKZE before 65 is taxed as income at the owner's marginal PIT rate, a setting the owner enters once. Until it is set, an IKZE's liquidation value is shown as missing rather than guessed.

docs/DESIGN.md#D-78
