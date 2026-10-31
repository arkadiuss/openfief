---
title: D-78's fee-rule minimum must be an amount in any currency, converted at the valuation day's rate.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-129-1
  - lane:strict
text_ref: design-decisions:D-129#1
origin:
  kind: import
  ref: design-decisions:D-129#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: D-78's fee-rule minimum must be an amount in any currency, converted at the valuation day's rate.
logic_claims:
  - CLAIM-3CA89A425F54721E
semantic_clauses:
  - D-78's fee-rule minimum must be an amount in any currency, converted at the valuation day's rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 95127f12b85a3c386f1412c87472984b8bea01ae12e0f31438847677fd8c8db4
semantic_inventory:
  - claim_key: CLAIM-3CA89A425F54721E
    claim_text: D-78's fee-rule minimum must be an amount in any currency, converted at the valuation day's rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
id: REQ-valuation-sale-fee-minimum-any-currency
type: req
---
D-78's fee-rule minimum must be an amount in any currency, converted at the valuation day's rate.

## Context

D-78 estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency (REQ-valuation-sale-fee-rule). It didn't say which currency the minimum is in. The implementation raised this as question #16 on docs/implementation-questions.md and built the minimum as an amount in any currency, converted at the FX rate of the valuation's day, a missing rate leaving the liquidation value missing. On 2026-10-31 the author confirmed that as D-129, rather than tying the minimum to the instrument's currency.

## Source

> D-78's minimum is an amount in any currency, converted at the valuation day's rate; the commission applies to holdings priced at market, not to bonds or cash; foreign means priced in another currency than the pack's; and the minimum applies to the commission before the currency-conversion percentage is added.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-129#1
