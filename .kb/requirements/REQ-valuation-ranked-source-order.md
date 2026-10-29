---
title: A kind's source ranking must list the sources the owner placed first, in the owner's order, then every other source that delivered that kind, the one naming the kind's first-party institution first and the rest by name, and a ranking that names a source twice or one with an empty or overlong name must be refused.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: A kind's source ranking must list the sources the owner placed first, in the owner's order, then every other source that delivered that kind, the one naming the kind's first-party institution first and the rest by name, and a ranking that names a source twice or one with an empty or overlong name must be refused.
logic_claims:
  - CLAIM-EEE0C143DCA8312F
semantic_clauses:
  - A kind's source ranking must list the sources the owner placed first, in the owner's order, then every other source that delivered that kind, the one naming the kind's first-party institution first and the rest by name, and a ranking that names a source twice or one with an empty or overlong name must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8f1768b041c080ad6157fa8cb39e14fae3d04a41f1f1022ca39cfd9717937c7e
semantic_inventory:
  - claim_key: CLAIM-EEE0C143DCA8312F
    claim_text: A kind's source ranking must list the sources the owner placed first, in the owner's order, then every other source that delivered that kind, the one naming the kind's first-party institution first and the rest by name, and a ranking that names a source twice or one with an empty or overlong name must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 313
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:08.667Z'
id: REQ-valuation-ranked-source-order
type: req
---
A kind's source ranking must list the sources the owner placed first, in the owner's order, then every other source that delivered that kind, the one naming the kind's first-party institution first and the rest by name, and a ranking that names a source twice or one with an empty or overlong name must be refused.

## Context

D-94 lets the owner rank sources and ranks the first-party connector first by default. A source is a connector instance and its default place comes from the institution it names, the narrowest reading (implementation questions). Modeled by Claude Code for the market-data work package (D-94, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-94]** The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
