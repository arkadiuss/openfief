---
title: Every price from a source other than its instrument's series source must be compared with the series source's latest price on or before its date like a cross-check price, and never value a holding.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: Every price from a source other than its instrument's series source must be compared with the series source's latest price on or before its date like a cross-check price, and never value a holding.
logic_claims:
  - CLAIM-C9CD77A25D0B21EB
semantic_clauses:
  - Every price from a source other than its instrument's series source must be compared with the series source's latest price on or before its date like a cross-check price, and never value a holding.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 84e4c407b7cdce190e8ed269648261b4b7723ebf06fa6e08667598762d789a47
semantic_inventory:
  - claim_key: CLAIM-C9CD77A25D0B21EB
    claim_text: Every price from a source other than its instrument's series source must be compared with the series source's latest price on or before its date like a cross-check price, and never value a holding
    role: normative
    status: modeled
    span:
      start: 0
      end: 196
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:59.911Z'
id: REQ-valuation-other-source-price-checks
type: req
---
Every price from a source other than its instrument's series source must be compared with the series source's latest price on or before its date like a cross-check price, and never value a holding.

## Context

D-94 stores every other source's values as cross-checks (D-22). The earlier refinement compared lower-ranked prices with the winner of each date; with one source per series every other source's price checks the series source's.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
