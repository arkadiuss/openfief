---
title: A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in PLN.
status: closed
text_ref: docs/DESIGN.md#D-128
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-128
  - lane:strict
semantic_text: A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in PLN.
logic_claims:
  - CLAIM-506D648697AD2E98
semantic_clauses:
  - A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in PLN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3caa4669d18372501baf3e3e7d6a82bd2be16a9c24bbb1ea6c3ec5b015494158
semantic_inventory:
  - claim_key: CLAIM-506D648697AD2E98
    claim_text: A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in PLN
    role: normative
    status: modeled
    span:
      start: 0
      end: 127
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:25.256Z'
id: REQ-valuation-stored-cost-basis
type: req
---
A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in PLN.

## Context

Totals offset gains on the whole (D-128), so they need each row's gain, not only its own tax. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-128]** A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

docs/DESIGN.md#D-128
