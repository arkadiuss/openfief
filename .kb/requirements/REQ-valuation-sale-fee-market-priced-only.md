---
title: D-78's sale commission must apply to holdings priced at market, not to bonds or cash.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-129-2
  - lane:strict
text_ref: design-decisions:D-129#2
origin:
  kind: import
  ref: design-decisions:D-129#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: D-78's sale commission must apply to holdings priced at market, not to bonds or cash.
logic_claims:
  - CLAIM-1044899C1A3B0652
semantic_clauses:
  - D-78's sale commission must apply to holdings priced at market, not to bonds or cash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7631666da623a6d1491fbcf51d3d99011333be39c2f404ad2dc462db1c9fa9f1
semantic_inventory:
  - claim_key: CLAIM-1044899C1A3B0652
    claim_text: D-78's sale commission must apply to holdings priced at market, not to bonds or cash
    role: normative
    status: modeled
    span:
      start: 0
      end: 84
id: REQ-valuation-sale-fee-market-priced-only
type: req
---
D-78's sale commission must apply to holdings priced at market, not to bonds or cash.

## Context

D-78's fee rule estimates a sale's commission in the liquidation value (REQ-valuation-sale-fee-rule), but didn't say whether bonds, whose redemption costs no commission, pay it. The implementation raised this as question #16 on docs/implementation-questions.md and applies the commission to holdings priced at market only, not to bonds or cash; the early-redemption fee covers a bond's exit cost instead. On 2026-10-31 the author confirmed that as D-129.

## Source

> D-78's minimum is an amount in any currency, converted at the valuation day's rate; the commission applies to holdings priced at market, not to bonds or cash; foreign means priced in another currency than the pack's; and the minimum applies to the commission before the currency-conversion percentage is added.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-129#2
