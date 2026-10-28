---
title: Market index levels must be stored per index name and date with a positive level, replace an earlier level of the same index and date, enqueue no revaluation, and be read back for one index through a date in date order.
status: open
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
semantic_text: Market index levels must be stored per index name and date with a positive level, replace an earlier level of the same index and date, enqueue no revaluation, and be read back for one index through a date in date order.
logic_claims:
  - CLAIM-9EE9F8A18693D363
semantic_clauses:
  - Market index levels must be stored per index name and date with a positive level, replace an earlier level of the same index and date, enqueue no revaluation, and be read back for one index through a date in date order.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bf6b17e17fe562f2a3d0f496cf0a4b888355f95101d112ac30045b453ca25371
semantic_inventory:
  - claim_key: CLAIM-9EE9F8A18693D363
    claim_text: Market index levels must be stored per index name and date with a positive level, replace an earlier level of the same index and date, enqueue no revaluation, and be read back for one index through a date in date order
    role: normative
    status: modeled
    span:
      start: 0
      end: 218
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:28.752Z'
id: REQ-valuation-index-levels
type: req
---
Market index levels must be stored per index name and date with a positive level, replace an earlier level of the same index and date, enqueue no revaluation, and be read back for one index through a date in date order.

## Context

D-51 brings end-of-day index history; D-52 lets a model target be an index, so its backtest (D-53) needs the levels stored. No holding is valued from an index, so nothing is revalued. Asked for by the orchestrator. Modeled by Claude Code for the model portfolio task (index targets need stored levels); the author has not reviewed it yet.

## Source

> **[D-51]** The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use.

docs/DESIGN.md#D-51

> **[D-52]** each target is an instrument (such as an index ETF or a stock) or an index, with a weight

docs/DESIGN.md#D-52
