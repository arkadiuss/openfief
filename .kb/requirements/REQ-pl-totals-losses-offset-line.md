---
title: A losses-offset line must reconcile the holding rows with the account and portfolio totals.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-128-3
  - lane:strict
text_ref: design-decisions:D-128#3
origin:
  kind: import
  ref: design-decisions:D-128#3
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: A losses-offset line must reconcile the holding rows with the account and portfolio totals.
logic_claims:
  - CLAIM-811CBB4522F0DC1E
semantic_clauses:
  - A losses-offset line must reconcile the holding rows with the account and portfolio totals.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: df319671f273e8f1383fd1516c999edec7725e5a4f006b661e142a6253a54a54
semantic_inventory:
  - claim_key: CLAIM-811CBB4522F0DC1E
    claim_text: A losses-offset line must reconcile the holding rows with the account and portfolio totals
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
id: REQ-pl-totals-losses-offset-line
type: req
---
A losses-offset line must reconcile the holding rows with the account and portfolio totals.

## Context

Under D-128 each holding's row carries its own tax as if it alone were sold (REQ-valuation-holding-tax-as-if-sold-alone), while account and portfolio totals net the tax year's gains and losses as PIT-38 does and round to whole złoty (REQ-pl-totals-tax-pit38), so the rows no longer add up to the total. The author chose both readings on 2026-10-31, answering question #15 on docs/implementation-questions.md, and added a losses-offset line so the owner sees why the total differs from the sum of the rows.

## Source

> A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-128#3
