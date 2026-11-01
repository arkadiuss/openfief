---
title: 'Every total the read API serves must count its net and liquidation values with the pack''s totals over its accounts'' rows, sales and payments, and report the losses-offset line: the total less its rows, or less the accounts'' own totals for the portfolio.'
status: open
text_ref: docs/DESIGN.md#D-128
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-128
  - lane:strict
semantic_text: 'Every total the read API serves must count its net and liquidation values with the pack''s totals over its accounts'' rows, sales and payments, and report the losses-offset line: the total less its rows, or less the accounts'' own totals for the portfolio.'
logic_claims:
  - CLAIM-3F8B1683EA3E5FBC
semantic_clauses:
  - 'Every total the read API serves must count its net and liquidation values with the pack''s totals over its accounts'' rows, sales and payments, and report the losses-offset line: the total less its rows, or less the accounts'' own totals for the portfolio.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 724b4dd979a47c1eb81578d1932ab8374a6b28a95857125a086808b524c796cd
semantic_inventory:
  - claim_key: CLAIM-3F8B1683EA3E5FBC
    claim_text: 'Every total the read API serves must count its net and liquidation values with the pack''s totals over its accounts'' rows, sales and payments, and report the losses-offset line: the total less its rows, or less the accounts'' own totals for the portfolio'
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:48.458Z'
id: REQ-views-totals-losses-offset
type: req
---
Every total the read API serves must count its net and liquidation values with the pack's totals over its accounts' rows, sales and payments, and report the losses-offset line: the total less its rows, or less the accounts' own totals for the portfolio.

## Context

D-128's losses-offset line reconciles the rows with the total. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-128]** A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

docs/DESIGN.md#D-128
