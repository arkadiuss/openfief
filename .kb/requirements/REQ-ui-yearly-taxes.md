---
title: The built-in UI must show the owner this year's yearly taxes per exemption class with the rate and tax, a form to classify each instrument counted without a class, and the losses-offset line under the accounts and on an account's value.
status: open
text_ref: docs/DESIGN.md#D-128
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-128
  - lane:strict
semantic_text: The built-in UI must show the owner this year's yearly taxes per exemption class with the rate and tax, a form to classify each instrument counted without a class, and the losses-offset line under the accounts and on an account's value.
logic_claims:
  - CLAIM-F4B79BB16C6EE892
semantic_clauses:
  - The built-in UI must show the owner this year's yearly taxes per exemption class with the rate and tax, a form to classify each instrument counted without a class, and the losses-offset line under the accounts and on an account's value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8e44f608f1e81f010071aa0d27533b22d6e1b614e2507cf72aad99c24115095a
semantic_inventory:
  - claim_key: CLAIM-F4B79BB16C6EE892
    claim_text: The built-in UI must show the owner this year's yearly taxes per exemption class with the rate and tax, a form to classify each instrument counted without a class, and the losses-offset line under the accounts and on an account's value
    role: normative
    status: modeled
    span:
      start: 0
      end: 235
origin:
  kind: agent
  recorded_at: '2026-10-31T18:35:02.101Z'
id: REQ-ui-yearly-taxes
type: req
---
The built-in UI must show the owner this year's yearly taxes per exemption class with the rate and tax, a form to classify each instrument counted without a class, and the losses-offset line under the accounts and on an account's value.

## Context

D-127, D-128 and D-132 ask for the OKI tax, the losses-offset line and the owner's classification. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-128]** A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

docs/DESIGN.md#D-128
