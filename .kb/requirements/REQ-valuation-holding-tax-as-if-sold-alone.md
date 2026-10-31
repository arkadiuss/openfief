---
title: A holding's net and liquidation values must carry its own tax as if it alone were sold.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-128-1
  - lane:strict
text_ref: design-decisions:D-128#1
origin:
  kind: import
  ref: design-decisions:D-128#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: A holding's net and liquidation values must carry its own tax as if it alone were sold.
logic_claims:
  - CLAIM-BE1418A3E4CD5EE3
semantic_clauses:
  - A holding's net and liquidation values must carry its own tax as if it alone were sold.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: dd35df7553eaf0961c304aba92d40ca07899ccd1d7c82fd2d027502f712e4fab
semantic_inventory:
  - claim_key: CLAIM-BE1418A3E4CD5EE3
    claim_text: A holding's net and liquidation values must carry its own tax as if it alone were sold
    role: normative
    status: modeled
    span:
      start: 0
      end: 86
id: REQ-valuation-holding-tax-as-if-sold-alone
type: req
---
A holding's net and liquidation values must carry its own tax as if it alone were sold.

## Context

D-77 gives every holding, account and total a net and a liquidation value after the taxes due (REQ-valuation-net-value, REQ-valuation-liquidation-value), but Polish tax counts gains per tax year over all disposals (PIT-38), rounds to whole złoty, and taxes an IKE's early return on what is withdrawn above what was paid in. The implementation raised this as question #15 on docs/implementation-questions.md: it taxes each holding on its own gain over its open lots' cost, and a loss offsets nothing. On 2026-10-31 the author chose both readings, recorded as D-128: each holding's row keeps its own tax as if it alone were sold, which is what the implementation already builds, while account and portfolio totals count the tax as PIT-38 does (REQ-pl-totals-tax-pit38).

## Source

> A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-128#1
