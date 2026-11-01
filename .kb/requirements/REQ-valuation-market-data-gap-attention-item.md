---
title: A gap in a series' values longer than a threshold the owner can set must raise an attention item.
status: open
text_ref: design-decisions:D-94#8
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-8
  - lane:strict
semantic_text: A gap in a series' values longer than a threshold the owner can set must raise an attention item.
logic_claims:
  - CLAIM-B38177D9D1FA241B
semantic_clauses:
  - A gap in a series' values longer than a threshold the owner can set must raise an attention item.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b0fc9aa9dcbe2ca90550dac88416889f3b3c6224ee6b87b7226af78b7f9560cb
semantic_inventory:
  - claim_key: CLAIM-B38177D9D1FA241B
    claim_text: A gap in a series' values longer than a threshold the owner can set must raise an attention item
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#8
  recorded_at: '2026-11-01T07:37:23.420Z'
id: REQ-valuation-market-data-gap-attention-item
type: req
---
A gap in a series' values longer than a threshold the owner can set must raise an attention item.

## Context

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. Carrying the latest earlier value forward hides a source that stopped delivering, so a gap longer than the owner's threshold raises an attention item in D-130's store (REQ-core-attention-item-store). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#8
