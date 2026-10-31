---
title: D-78's minimum must apply to the commission before the currency-conversion percentage is added.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-129-4
  - lane:strict
text_ref: design-decisions:D-129#4
origin:
  kind: import
  ref: design-decisions:D-129#4
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: D-78's minimum must apply to the commission before the currency-conversion percentage is added.
logic_claims:
  - CLAIM-11F1D3418E7F474C
semantic_clauses:
  - D-78's minimum must apply to the commission before the currency-conversion percentage is added.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0655dddf53665e812b8ba39a53938f59b1a4487506b79b83a93ebf06baee5baa
semantic_inventory:
  - claim_key: CLAIM-11F1D3418E7F474C
    claim_text: D-78's minimum must apply to the commission before the currency-conversion percentage is added
    role: normative
    status: modeled
    span:
      start: 0
      end: 94
id: REQ-valuation-sale-fee-minimum-before-conversion
type: req
---
D-78's minimum must apply to the commission before the currency-conversion percentage is added.

## Context

D-78's fee rule is a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency (REQ-valuation-sale-fee-rule), and didn't say whether the minimum covers the conversion percentage too. The implementation raised the fee rule's details as question #16 on docs/implementation-questions.md: the minimum applies to the commission alone, before the conversion percentage is added (REQ-jurisdiction-fee-rule-sale-cost, which the author has not reviewed). On 2026-10-31 the author confirmed that as D-129.

## Source

> D-78's minimum is an amount in any currency, converted at the valuation day's rate; the commission applies to holdings priced at market, not to bonds or cash; foreign means priced in another currency than the pack's; and the minimum applies to the commission before the currency-conversion percentage is added.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-129#4
