---
title: Valuation totals must sum, per date, the valued holdings and each account's cash, converted at the FX rate on or before the date and valued in the account's tax wrapper, with cash that has no rate counted only per currency.
status: open
text_ref: docs/DESIGN.md#D-106
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-106
  - lane:strict
semantic_text: Valuation totals must sum, per date, the valued holdings and each account's cash, converted at the FX rate on or before the date and valued in the account's tax wrapper, with cash that has no rate counted only per currency.
logic_claims:
  - CLAIM-8B1F795085E497B2
semantic_clauses:
  - Valuation totals must sum, per date, the valued holdings and each account's cash, converted at the FX rate on or before the date and valued in the account's tax wrapper, with cash that has no rate counted only per currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c6b039639367c974071f1db6862c5fde2ee11268363b4bc3178fa1fc5d5222a8
semantic_inventory:
  - claim_key: CLAIM-8B1F795085E497B2
    claim_text: Valuation totals must sum, per date, the valued holdings and each account's cash, converted at the FX rate on or before the date and valued in the account's tax wrapper, with cash that has no rate counted only per currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 222
origin:
  kind: agent
  recorded_at: '2026-10-30T00:05:13.296Z'
id: REQ-views-totals-with-cash
type: req
---
Valuation totals must sum, per date, the valued holdings and each account's cash, converted at the FX rate on or before the date and valued in the account's tax wrapper, with cash that has no rate counted only per currency.

## Context

Supersedes REQ-views-valuation-totals, which summed stored valuations only: D-106 makes valuation totals include cash. Valuing cash in its wrapper follows D-77. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-views-valuation-totals: D-106 makes valuation totals include cash. The superseded requirement is closed.

## Source

> **[D-106]** Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation value. [...] Valuation totals include cash.

docs/DESIGN.md#D-106
