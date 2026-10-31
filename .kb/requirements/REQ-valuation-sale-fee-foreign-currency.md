---
title: In D-78's fee rule, foreign must mean priced in another currency than the jurisdiction pack's.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-129-3
  - lane:strict
text_ref: design-decisions:D-129#3
origin:
  kind: import
  ref: design-decisions:D-129#3
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: In D-78's fee rule, foreign must mean priced in another currency than the jurisdiction pack's.
logic_claims:
  - CLAIM-210DD8DA948FFB36
semantic_clauses:
  - In D-78's fee rule, foreign must mean priced in another currency than the jurisdiction pack's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c1c25e47f23ca32afa9917e2dead3e4e73a8ea11984b9391eb31d20cdf17b885
semantic_inventory:
  - claim_key: CLAIM-210DD8DA948FFB36
    claim_text: In D-78's fee rule, foreign must mean priced in another currency than the jurisdiction pack's
    role: normative
    status: modeled
    span:
      start: 0
      end: 93
id: REQ-valuation-sale-fee-foreign-currency
type: req
---
In D-78's fee rule, foreign must mean priced in another currency than the jurisdiction pack's.

## Context

D-78's fee rule adds a currency-conversion percentage for instruments in a foreign currency (REQ-valuation-sale-fee-rule), but didn't say what foreign is relative to. The implementation raised this as question #16 on docs/implementation-questions.md and reads foreign as priced in another currency than the pack's, PLN for Poland. On 2026-10-31 the author confirmed that as D-129, rather than charging the conversion percentage on foreign cash too.

## Source

> D-78's minimum is an amount in any currency, converted at the valuation day's rate; the commission applies to holdings priced at market, not to bonds or cash; foreign means priced in another currency than the pack's; and the minimum applies to the commission before the currency-conversion percentage is added.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-129#3
