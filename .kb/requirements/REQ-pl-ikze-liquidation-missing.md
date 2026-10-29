---
title: Until the owner's marginal PIT rate is set, an IKZE's liquidation value must be shown as missing rather than guessed.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-78-4
  - lane:strict
text_ref: design-decisions:D-78#4
origin:
  kind: import
  ref: design-decisions:D-78#4
  recorded_at: '2026-10-29T08:04:37.937Z'
semantic_text: Until the owner's marginal PIT rate is set, an IKZE's liquidation value must be shown as missing rather than guessed.
logic_claims:
  - CLAIM-0F0AFDAA0DC98E25
semantic_clauses:
  - Until the owner's marginal PIT rate is set, an IKZE's liquidation value must be shown as missing rather than guessed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4a519233b2ab69c5ae66a37bb3c8eb0004ea215813c61dc4d019fd80431467e0
semantic_inventory:
  - claim_key: CLAIM-0F0AFDAA0DC98E25
    claim_text: Until the owner's marginal PIT rate is set, an IKZE's liquidation value must be shown as missing rather than guessed
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-pl-ikze-liquidation-missing
type: req
---
Until the owner's marginal PIT rate is set, an IKZE's liquidation value must be shown as missing rather than guessed.

## Context

D-78 taxes an IKZE withdrawal before 65 as income at the owner's marginal PIT rate, a setting the owner enters once (REQ-pl-ikze-early-withdrawal-tax), and D-77's liquidation value deducts that tax (REQ-valuation-liquidation-value). Without the rate the tax can't be known, so the value is shown as missing; a sale fee rule that was never set, by contrast, counts as zero (REQ-valuation-sale-fee-rule-zero). The implementation already records a missing input instead of a value when a price, FX rate or index reading is absent (REQ-valuation-missing-input). The author recorded D-76 to D-78 after the design session that answered the implementation's questions.

## Source

> Liquidation value estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency. The rule is zero until set. Withdrawing from IKZE before 65 is taxed as income at the owner's marginal PIT rate, a setting the owner enters once. Until it is set, an IKZE's liquidation value is shown as missing rather than guessed.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-78#4
