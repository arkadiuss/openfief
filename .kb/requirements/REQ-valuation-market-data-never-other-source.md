---
title: A series' value must never come from another source on the days its chosen source lacks one.
status: open
text_ref: design-decisions:D-94#9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-9
  - lane:strict
semantic_text: A series' value must never come from another source on the days its chosen source lacks one.
logic_claims:
  - CLAIM-E626CA37756CA0CA
semantic_clauses:
  - A series' value must never come from another source on the days its chosen source lacks one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0eccd0aa4dbe537db87afcc773893afacb29df44055dd9cab14d8a00a7ac62c2
semantic_inventory:
  - claim_key: CLAIM-E626CA37756CA0CA
    claim_text: A series' value must never come from another source on the days its chosen source lacks one
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#9
  recorded_at: '2026-11-01T07:37:30.216Z'
id: REQ-valuation-market-data-never-other-source
type: req
---
A series' value must never come from another source on the days its chosen source lacks one.

## Context

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. Filling a missing day from another source is exactly the silent switch D-94 rules out; a missing day takes the latest earlier value instead (REQ-valuation-market-data-missing-day-carries-forward). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#9
