---
title: A day without a value from a series' source must take the latest earlier value.
status: open
text_ref: design-decisions:D-94#7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-7
  - lane:strict
semantic_text: A day without a value from a series' source must take the latest earlier value.
logic_claims:
  - CLAIM-CF87DCEBE225DD42
semantic_clauses:
  - A day without a value from a series' source must take the latest earlier value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1904bf9f79dc6e07ed858c61548083f0c4d5ae2c1ba381802c28a35e4d0207b0
semantic_inventory:
  - claim_key: CLAIM-CF87DCEBE225DD42
    claim_text: A day without a value from a series' source must take the latest earlier value
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#7
  recorded_at: '2026-11-01T07:37:16.098Z'
id: REQ-valuation-market-data-missing-day-carries-forward
type: req
---
A day without a value from a series' source must take the latest earlier value.

## Context

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. Markets close on holidays and sources skip days; the series carries its latest earlier value from the same source instead of taking one from another source (REQ-valuation-market-data-never-other-source). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#7
