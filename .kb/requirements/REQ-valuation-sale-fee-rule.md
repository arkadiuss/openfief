---
title: 'The liquidation value must estimate a sale''s commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-78-1
  - lane:strict
text_ref: design-decisions:D-78#1
origin:
  kind: import
  ref: design-decisions:D-78#1
  recorded_at: '2026-10-29T08:02:24.227Z'
semantic_text: 'The liquidation value must estimate a sale''s commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency.'
logic_claims:
  - CLAIM-D69A967F5E1A1BFD
semantic_clauses:
  - 'The liquidation value must estimate a sale''s commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c697ad1d6934945cf15e25a2db527946e4752278edbf075e4ebfc4d7899af547
semantic_inventory:
  - claim_key: CLAIM-D69A967F5E1A1BFD
    claim_text: 'The liquidation value must estimate a sale''s commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency'
    role: normative
    status: modeled
    span:
      start: 0
      end: 204
id: REQ-valuation-sale-fee-rule
type: req
---
The liquidation value must estimate a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency.

## Context

D-77's liquidation value is the cash in hand after every tax and fee due today, sale commissions included (REQ-valuation-liquidation-value); D-78 says how the commission is estimated, and REQ-valuation-sale-fee-rule-zero records that the rule is zero until set. In M0 fees go to the fee ledger account and are not added to lot cost (D-34). The author recorded D-76 to D-78 after the design session that answered the implementation's questions; no further reason is stated.

## Source

> Liquidation value estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency. The rule is zero until set. Withdrawing from IKZE before 65 is taxed as income at the owner's marginal PIT rate, a setting the owner enters once. Until it is set, an IKZE's liquidation value is shown as missing rather than guessed.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-78#1
