---
title: Withdrawing from IKZE before 65 must be taxed as income at the owner's marginal PIT rate, a setting the owner enters once.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-78-3
  - lane:strict
text_ref: design-decisions:D-78#3
origin:
  kind: import
  ref: design-decisions:D-78#3
  recorded_at: '2026-10-29T08:03:56.208Z'
semantic_text: Withdrawing from IKZE before 65 must be taxed as income at the owner's marginal PIT rate, a setting the owner enters once.
logic_claims:
  - CLAIM-023DA4914B167148
semantic_clauses:
  - Withdrawing from IKZE before 65 must be taxed as income at the owner's marginal PIT rate, a setting the owner enters once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2cb31cdc60be1bb8daff431ee6618407ad0a3d1fe8e7873b4dbb9479021a1410
semantic_inventory:
  - claim_key: CLAIM-023DA4914B167148
    claim_text: Withdrawing from IKZE before 65 must be taxed as income at the owner's marginal PIT rate, a setting the owner enters once
    role: normative
    status: modeled
    span:
      start: 0
      end: 121
id: REQ-pl-ikze-early-withdrawal-tax
type: req
---
Withdrawing from IKZE before 65 must be taxed as income at the owner's marginal PIT rate, a setting the owner enters once.

## Context

IKZE is one of the Polish pack's tax wrappers under D-76 (REQ-pl-tax-wrappers). D-77's liquidation value deducts early-withdrawal taxes due today (REQ-valuation-liquidation-value), and D-78 says how this one is computed; until the owner enters the rate, an IKZE's liquidation value is missing (REQ-pl-ikze-liquidation-missing). The implementation deducts no Belka tax inside IKZE and leaves the IKZE withdrawal tax to the account (REQ-pl-tax-wrapper-exempt). The author recorded D-76 to D-78 after the design session that answered the implementation's questions; no further reason is stated.

## Source

> Liquidation value estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency. The rule is zero until set. Withdrawing from IKZE before 65 is taxed as income at the owner's marginal PIT rate, a setting the owner enters once. Until it is set, an IKZE's liquidation value is shown as missing rather than guessed.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-78#3
