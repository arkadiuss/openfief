---
title: An account's sale fee rule must be zero until the owner sets it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-78-2
  - lane:strict
text_ref: design-decisions:D-78#2
origin:
  kind: import
  ref: design-decisions:D-78#2
  recorded_at: '2026-10-29T08:03:06.308Z'
semantic_text: An account's sale fee rule must be zero until the owner sets it.
logic_claims:
  - CLAIM-9B2B8F6BC2A8BD5B
semantic_clauses:
  - An account's sale fee rule must be zero until the owner sets it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e08264a8da8b5f2d4bc1fdb7b3ffcf9d7a1942890ccc527bf8596ae720e9277e
semantic_inventory:
  - claim_key: CLAIM-9B2B8F6BC2A8BD5B
    claim_text: An account's sale fee rule must be zero until the owner sets it
    role: normative
    status: modeled
    span:
      start: 0
      end: 63
id: REQ-valuation-sale-fee-rule-zero
type: req
---
An account's sale fee rule must be zero until the owner sets it.

## Context

D-78 has the liquidation value estimate a sale's commission with a fee rule the owner sets per account (REQ-valuation-sale-fee-rule), and D-77's liquidation value deducts sale commissions (REQ-valuation-liquidation-value). Unlike the IKZE marginal PIT rate, whose absence leaves an IKZE's liquidation value missing (REQ-pl-ikze-liquidation-missing), a fee rule that was never set counts as zero. The author recorded D-76 to D-78 after the design session that answered the implementation's questions; no further reason is stated.

## Source

> Liquidation value estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency. The rule is zero until set. Withdrawing from IKZE before 65 is taxed as income at the owner's marginal PIT rate, a setting the owner enters once. Until it is set, an IKZE's liquidation value is shown as missing rather than guessed.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-78#2
