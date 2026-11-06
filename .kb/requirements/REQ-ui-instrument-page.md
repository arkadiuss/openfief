---
title: An instrument's page must show its display name before its tickers and ISIN, a breadcrumb through the account it was opened from, each listing's latest price and a year of closing prices per source with the primary one marked, its holdings and open lots and its transactions by account, its price checks with their history and the owner's forms returning to the page, its yearly tax classes and its open attention items.
status: open
text_ref: docs/DESIGN.md#D-195
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-195
  - lane:strict
semantic_text: An instrument's page must show its display name before its tickers and ISIN, a breadcrumb through the account it was opened from, each listing's latest price and a year of closing prices per source with the primary one marked, its holdings and open lots and its transactions by account, its price checks with their history and the owner's forms returning to the page, its yearly tax classes and its open attention items.
logic_claims:
  - CLAIM-344125E3C9D1CEB4
semantic_clauses:
  - An instrument's page must show its display name before its tickers and ISIN, a breadcrumb through the account it was opened from, each listing's latest price and a year of closing prices per source with the primary one marked, its holdings and open lots and its transactions by account, its price checks with their history and the owner's forms returning to the page, its yearly tax classes and its open attention items.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6d09b2e17c36d0271f64420c6cae9d2130e78c862a3738ef1acd0e9605ffe60c
semantic_inventory:
  - claim_key: CLAIM-344125E3C9D1CEB4
    claim_text: An instrument's page must show its display name before its tickers and ISIN, a breadcrumb through the account it was opened from, each listing's latest price and a year of closing prices per source with the primary one marked, its holdings and open lots and its transactions by account, its price checks with their history and the owner's forms returning to the page, its yearly tax classes and its open attention items
    role: normative
    status: modeled
    span:
      start: 0
      end: 419
origin:
  kind: agent
  recorded_at: '2026-11-06T10:43:18.327Z'
id: REQ-ui-instrument-page
type: req
---
An instrument's page must show its display name before its tickers and ISIN, a breadcrumb through the account it was opened from, each listing's latest price and a year of closing prices per source with the primary one marked, its holdings and open lots and its transactions by account, its price checks with their history and the owner's forms returning to the page, its yearly tax classes and its open attention items.

## Context

Refines D-195 and D-197 from the design session's mockups. Modeled by Claude Code for the UI v2 pages slice; the author has not reviewed the wording.

## Source

> **[D-195]** Every figure links down to what it is made of: the total to its accounts, an account to its holdings, a holding to its instrument, an allocation bar to the holdings behind it. An instrument has its own page: its price history and source, its holdings and lots across accounts, its transactions, its price checks, its ticker and ISIN, and its OKI class.

docs/DESIGN.md#D-195

> **[D-197]** An instrument has a display name: the first one a connector delivers for it, such as XTB's display name, IBKR's description or the market-data lookup's, which the owner can change. Pages show it before the ticker and ISIN.

docs/DESIGN.md#D-197
