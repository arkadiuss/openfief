---
title: A connector plugin must write an end-of-day price as its instrument's ISIN and venue symbol, its date and its close with currency, and a market index level as the index's name, its date and its level, as exact decimal strings and YYYY-MM-DD dates.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: A connector plugin must write an end-of-day price as its instrument's ISIN and venue symbol, its date and its close with currency, and a market index level as the index's name, its date and its level, as exact decimal strings and YYYY-MM-DD dates.
logic_claims:
  - CLAIM-C7564A3EC83ACA2A
semantic_clauses:
  - A connector plugin must write an end-of-day price as its instrument's ISIN and venue symbol, its date and its close with currency, and a market index level as the index's name, its date and its level, as exact decimal strings and YYYY-MM-DD dates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 78388917a929b0cc6390b4f1a31ba8e8c89a542e77ee34b5d39b2f2a6af9d7b4
semantic_inventory:
  - claim_key: CLAIM-C7564A3EC83ACA2A
    claim_text: A connector plugin must write an end-of-day price as its instrument's ISIN and venue symbol, its date and its close with currency, and a market index level as the index's name, its date and its level, as exact decimal strings and YYYY-MM-DD dates
    role: normative
    status: modeled
    span:
      start: 0
      end: 246
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:57.186Z'
id: REQ-ingest-delivery-market-data
type: req
---
A connector plugin must write an end-of-day price as its instrument's ISIN and venue symbol, its date and its close with currency, and a market index level as the index's name, its date and its level, as exact decimal strings and YYYY-MM-DD dates.

## Context

The connector contract (D-49) carries prices and, since this change, index levels (message IndexLevel, Delivery field index_levels) for D-51's end-of-day index history; amounts are decimal strings (REQ-connector-contract-messages). Extends REQ-ingest-delivery-contract-messages to the market-data messages. Modeled by Claude Code for the EODHD market-data connector task; the author has not reviewed it yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49

> The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use. Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51
