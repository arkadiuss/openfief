---
title: The valuation store must return every stored price of an instrument's listings in a date range from every source, in listing, date and source order, and the latest stored price of each listing from each source.
status: open
text_ref: docs/DESIGN.md#D-195
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-195
  - lane:strict
semantic_text: The valuation store must return every stored price of an instrument's listings in a date range from every source, in listing, date and source order, and the latest stored price of each listing from each source.
logic_claims:
  - CLAIM-B34BE9C2BA4B9956
semantic_clauses:
  - The valuation store must return every stored price of an instrument's listings in a date range from every source, in listing, date and source order, and the latest stored price of each listing from each source.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c77e388e492c3cd20ada3fb9a5cad662ee75059cdbd2d480d40f82e83f398025
semantic_inventory:
  - claim_key: CLAIM-B34BE9C2BA4B9956
    claim_text: The valuation store must return every stored price of an instrument's listings in a date range from every source, in listing, date and source order, and the latest stored price of each listing from each source
    role: normative
    status: modeled
    span:
      start: 0
      end: 209
origin:
  kind: agent
  recorded_at: '2026-11-06T01:12:31.691Z'
id: REQ-valuation-instrument-price-history
type: req
---
The valuation store must return every stored price of an instrument's listings in a date range from every source, in listing, date and source order, and the latest stored price of each listing from each source.

## Context

D-195 gives an instrument its own page with its price history and source. D-94 keeps every ranked source's prices, so the history serves them all and the read API marks the one the listing is valued at. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-195]** ... An instrument has its own page: its price history and source, its holdings and lots across accounts, its transactions, its price checks, its ticker and ISIN, and its OKI class.

docs/DESIGN.md#D-195
