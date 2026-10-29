---
title: Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing.
logic_claims:
  - CLAIM-9C6C219D57E65B26
semantic_clauses:
  - Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e0878facc0ab048c4845893bec63c153798b2b3cd1f641b5c00f2fa6fdd79819
semantic_inventory:
  - claim_key: CLAIM-9C6C219D57E65B26
    claim_text: Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 226
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:14.359Z'
id: REQ-valuation-ranking-revalues
type: req
---
Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing.

## Context

D-36 revalues on any change to prices, FX rates or index readings; a new ranking changes which value counts on any day. Modeled by Claude Code for the market-data work package (D-94, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-94]** The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94

> **[D-36]** Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River).

docs/DESIGN.md#D-36
