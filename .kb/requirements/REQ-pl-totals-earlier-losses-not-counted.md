---
title: Losses carried from earlier tax years must not be counted in the totals' tax.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-128-4
  - lane:strict
text_ref: design-decisions:D-128#4
origin:
  kind: import
  ref: design-decisions:D-128#4
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Losses carried from earlier tax years must not be counted in the totals' tax.
logic_claims:
  - CLAIM-E4E373D78EE5FF66
semantic_clauses:
  - Losses carried from earlier tax years must not be counted in the totals' tax.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7e0e0417279c1c70c522fec29ee0ce9649a197233efaf299abbd8102dc27a0e1
semantic_inventory:
  - claim_key: CLAIM-E4E373D78EE5FF66
    claim_text: Losses carried from earlier tax years must not be counted in the totals' tax
    role: normative
    status: modeled
    span:
      start: 0
      end: 76
id: REQ-pl-totals-earlier-losses-not-counted
type: req
---
Losses carried from earlier tax years must not be counted in the totals' tax.

## Context

Under D-128 account and portfolio totals count the tax as PIT-38 does, with the tax year's gains and losses of every taxable account offsetting each other (REQ-pl-totals-tax-pit38). PIT-38 also lets a taxpayer deduct losses carried from earlier years. When the author decided D-128 on 2026-10-31, answering question #15 on docs/implementation-questions.md, they left those carried losses out of the totals; reason not stated.

## Source

> A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-128#4
