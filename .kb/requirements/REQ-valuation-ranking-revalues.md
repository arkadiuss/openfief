---
title: Changing the owner's ranking of a kind of market data must enqueue a revaluation of every account from its first transaction in the same database transaction that stores it, and saving an unchanged ranking must enqueue nothing.
status: closed
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

An earlier refinement of D-94 by Claude Code (default firmness).

## Why this changed

D-94 was rewritten: a series' source is chosen from the ranking only when the series first needs a value and is then kept, so a change of ranking changes no stored value and revalues nothing. Superseded by REQ-valuation-series-source-recorded.

## Source

docs/DESIGN.md#D-94
