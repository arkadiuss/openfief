---
title: The owner must be able to change the source of a single instrument, currency pair or index.
status: open
text_ref: design-decisions:D-94#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-6
  - lane:strict
semantic_text: The owner must be able to change the source of a single instrument, currency pair or index.
logic_claims:
  - CLAIM-5C453F842F9E1103
semantic_clauses:
  - The owner must be able to change the source of a single instrument, currency pair or index.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6a412cf84e4b12c76318c409aec24cb9c7862c970e779117cb3671a71fa80edb
semantic_inventory:
  - claim_key: CLAIM-5C453F842F9E1103
    claim_text: The owner must be able to change the source of a single instrument, currency pair or index
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-94#6
  recorded_at: '2026-11-01T07:37:08.752Z'
id: REQ-valuation-market-data-source-owner-switch
type: req
---
The owner must be able to change the source of a single instrument, currency pair or index.

## Context

On 2026-11-01 the author rewrote D-94 and D-124: each instrument, currency pair and index now has exactly one source, so that a series never switches source silently. The ranking chooses a series' source when it first needs a value; the owner can change that choice for one series without changing the ranking. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#6
