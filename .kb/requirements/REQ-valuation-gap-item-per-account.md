---
title: A revaluation must raise a market-data-gap attention item, its subject the account and the series, for each series a holding needs on today whose source's latest value is older than the gap threshold, and clear it for every other series it read.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: A revaluation must raise a market-data-gap attention item, its subject the account and the series, for each series a holding needs on today whose source's latest value is older than the gap threshold, and clear it for every other series it read.
logic_claims:
  - CLAIM-5D6BDA134D2707F2
semantic_clauses:
  - A revaluation must raise a market-data-gap attention item, its subject the account and the series, for each series a holding needs on today whose source's latest value is older than the gap threshold, and clear it for every other series it read.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3665f41af98e3aff676310c424e23d80578e118999e583810384c936f8951c14
semantic_inventory:
  - claim_key: CLAIM-5D6BDA134D2707F2
    claim_text: A revaluation must raise a market-data-gap attention item, its subject the account and the series, for each series a holding needs on today whose source's latest value is older than the gap threshold, and clear it for every other series it read
    role: normative
    status: modeled
    span:
      start: 0
      end: 244
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:50.429Z'
id: REQ-valuation-gap-item-per-account
type: req
---
A revaluation must raise a market-data-gap attention item, its subject the account and the series, for each series a holding needs on today whose source's latest value is older than the gap threshold, and clear it for every other series it read.

## Context

D-94 raises an attention item (D-130) for a gap longer than the threshold. Measuring it on today for the series a holding needs, one item per account and series so accounts never clear each other's, and a detail that stays the same while the gap lasts, are Claude Code's narrow reading, not reviewed by the author.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
