---
title: 'Account and portfolio totals must count the tax as PIT-38 does: the tax year''s realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-128-2
  - lane:strict
text_ref: design-decisions:D-128#2
origin:
  kind: import
  ref: design-decisions:D-128#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: 'Account and portfolio totals must count the tax as PIT-38 does: the tax year''s realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty.'
logic_claims:
  - CLAIM-FE46F0A32938E4F3
semantic_clauses:
  - 'Account and portfolio totals must count the tax as PIT-38 does: the tax year''s realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 09b3f2f7b5e2f6ad2d4e9c40974db2db53e4b845f75fdcd99524c4a119040153
semantic_inventory:
  - claim_key: CLAIM-FE46F0A32938E4F3
    claim_text: 'Account and portfolio totals must count the tax as PIT-38 does: the tax year''s realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty'
    role: normative
    status: modeled
    span:
      start: 0
      end: 216
id: REQ-pl-totals-tax-pit38
type: req
---
Account and portfolio totals must count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty.

## Context

D-77 gives every account and total a net and a liquidation value after the taxes due (REQ-valuation-three-values). The implementation raised how those taxes are counted as question #15 on docs/implementation-questions.md: it taxes each holding alone at 19%, rounded half up to grosze, a loss offsets nothing, and totals sum the holdings (REQ-views-totals-with-cash, which the author has not reviewed). Polish tax instead counts a tax year's gains over all disposals on PIT-38 and rounds the base and the tax to whole złoty. On 2026-10-31 the author chose both readings, recorded as D-128: holding rows keep their own tax (REQ-valuation-holding-tax-as-if-sold-alone), totals net gains and losses as PIT-38 does, and a losses-offset line reconciles the two (REQ-pl-totals-losses-offset-line).

## Source

> A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-128#2
