---
title: The read API must give the portfolio's value in PLN on a date grouped by asset class, currency, institution or sleeve, with each group's exact sum and the total, and must list the holdings it could not value instead of counting them.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: The read API must give the portfolio's value in PLN on a date grouped by asset class, currency, institution or sleeve, with each group's exact sum and the total, and must list the holdings it could not value instead of counting them.
logic_claims:
  - CLAIM-FDB7CF01DA11E298
semantic_clauses:
  - The read API must give the portfolio's value in PLN on a date grouped by asset class, currency, institution or sleeve, with each group's exact sum and the total, and must list the holdings it could not value instead of counting them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 15c38bce27fdb08df178d4eaffef3b95d366acd0fcefdf14ff21c7d4ce4d3a66
semantic_inventory:
  - claim_key: CLAIM-FDB7CF01DA11E298
    claim_text: The read API must give the portfolio's value in PLN on a date grouped by asset class, currency, institution or sleeve, with each group's exact sum and the total, and must list the holdings it could not value instead of counting them
    role: normative
    status: modeled
    span:
      start: 0
      end: 232
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:18.885Z'
id: REQ-views-allocation
type: req
---
The read API must give the portfolio's value in PLN on a date grouped by asset class, currency, institution or sleeve, with each group's exact sum and the total, and must list the holdings it could not value instead of counting them.

## Context

Refines the firm requirement REQ-AUTO-68636F42D98D4CF7: the built-in UI reads the record only through the public API (D-6), so the API must carry the allocation chart's numbers. D-30's broker is the glossary's institution. Asset classes are cash, treasury_bond (a Polish ISIN with a bond series code) and listed (anything else, a stock or an ETF in M0); a finer taxonomy needs reference data that is not decided. Groups use PLN so they can be added; the currency dimension groups by the original currency. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

docs/DESIGN.md D-30

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6
