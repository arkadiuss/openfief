---
title: The core must record the connector instance, with the institution its delivery names, as a source of each kind of market data the delivery carries, before storing the values.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The core must record the connector instance, with the institution its delivery names, as a source of each kind of market data the delivery carries, before storing the values.
logic_claims:
  - CLAIM-54BF0173B4ED98D2
semantic_clauses:
  - The core must record the connector instance, with the institution its delivery names, as a source of each kind of market data the delivery carries, before storing the values.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d191c482b48d7a89d08263360812896622ea526c0e1f601b8cf7da04a0f5e082
semantic_inventory:
  - claim_key: CLAIM-54BF0173B4ED98D2
    claim_text: The core must record the connector instance, with the institution its delivery names, as a source of each kind of market data the delivery carries, before storing the values
    role: normative
    status: modeled
    span:
      start: 0
      end: 173
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:36.770Z'
id: REQ-ingest-market-data-sources-recorded
type: req
---
The core must record the connector instance, with the institution its delivery names, as a source of each kind of market data the delivery carries, before storing the values.

## Context

The ranking needs to know which instances deliver each kind, and the institution sets the default place. Modeled by Claude Code for the market-data work package (D-94, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-94]** The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
