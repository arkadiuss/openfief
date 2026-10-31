---
title: An IKE's early return must be taxed on what is withdrawn above what was paid in.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-128-5
  - lane:strict
text_ref: design-decisions:D-128#5
origin:
  kind: import
  ref: design-decisions:D-128#5
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: An IKE's early return must be taxed on what is withdrawn above what was paid in.
logic_claims:
  - CLAIM-CC668C98B16C7A20
semantic_clauses:
  - An IKE's early return must be taxed on what is withdrawn above what was paid in.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4d306680f01212b77a103dcc3cb5158c95fd4a24d9089115ac59c3df6dac2e5f
semantic_inventory:
  - claim_key: CLAIM-CC668C98B16C7A20
    claim_text: An IKE's early return must be taxed on what is withdrawn above what was paid in
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
id: REQ-pl-ike-early-return-above-payments
type: req
---
An IKE's early return must be taxed on what is withdrawn above what was paid in.

## Context

D-77's liquidation value includes early-withdrawal taxes (REQ-valuation-liquidation-value). The implementation raised how they are counted as question #15 on docs/implementation-questions.md: it taxes an IKE's early return per holding like a taxable sale, so dividends and realized gains kept as cash in the IKE go untaxed in its liquidation value (REQ-pl-values-per-wrapper, which the author has not reviewed), while Polish tax charges a return from IKE on what is withdrawn above what was paid in. On 2026-10-31 the author decided, as D-128, that IKE and IKZE follow their own rules rather than PIT-38's netting, and that an IKE's early return is taxed on withdrawals above payments in. IKZE's early withdrawal stays taxed at the owner's marginal rate (REQ-pl-ikze-early-withdrawal-tax).

## Source

> A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-128#5
