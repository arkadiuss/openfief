---
title: 'The read API must serve an instrument by ISIN: its listings with price source and latest price, asset class and exemption classes, its prices from every source with the primary one marked, its holdings with lots and its transactions across the accounts the caller may view, and its price checks.'
status: open
text_ref: docs/DESIGN.md#D-195
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-195
  - lane:strict
semantic_text: 'The read API must serve an instrument by ISIN: its listings with price source and latest price, asset class and exemption classes, its prices from every source with the primary one marked, its holdings with lots and its transactions across the accounts the caller may view, and its price checks.'
logic_claims:
  - CLAIM-DA2B3A886F66120F
semantic_clauses:
  - 'The read API must serve an instrument by ISIN: its listings with price source and latest price, asset class and exemption classes, its prices from every source with the primary one marked, its holdings with lots and its transactions across the accounts the caller may view, and its price checks.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e940f3b515d54caa04b3d960c3ac85efaf2f0e72abacbc7e2088121df1de6de6
semantic_inventory:
  - claim_key: CLAIM-DA2B3A886F66120F
    claim_text: 'The read API must serve an instrument by ISIN: its listings with price source and latest price, asset class and exemption classes, its prices from every source with the primary one marked, its holdings with lots and its transactions across the accounts the caller may view, and its price checks'
    role: normative
    status: modeled
    span:
      start: 0
      end: 294
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:48.478Z'
id: REQ-views-instrument-page
type: req
---
The read API must serve an instrument by ISIN: its listings with price source and latest price, asset class and exemption classes, its prices from every source with the primary one marked, its holdings with lots and its transactions across the accounts the caller may view, and its price checks.

## Context

D-195 gives an instrument its own page: price history and source, holdings and lots across accounts, transactions, price checks, ticker and ISIN, OKI class. D-163 limits holdings and transactions to viewable accounts; prices are not tied to an account. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-195]** ... An instrument has its own page: its price history and source, its holdings and lots across accounts, its transactions, its price checks, its ticker and ISIN, and its OKI class.

docs/DESIGN.md#D-195

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account ...

docs/DESIGN.md#D-163
