---
title: A kind's gap threshold must be 7 days for prices and FX rates and 92 days for index readings unless the owner sets another, from 1 to 3660 days.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: A kind's gap threshold must be 7 days for prices and FX rates and 92 days for index readings unless the owner sets another, from 1 to 3660 days.
logic_claims:
  - CLAIM-5FDCD8ADCEBC5D53
semantic_clauses:
  - A kind's gap threshold must be 7 days for prices and FX rates and 92 days for index readings unless the owner sets another, from 1 to 3660 days.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1fdbbf8717de10ffef14df7f083edd6a4a681dd0b930c4d14f8cdce9fbcdb9fb
semantic_inventory:
  - claim_key: CLAIM-5FDCD8ADCEBC5D53
    claim_text: A kind's gap threshold must be 7 days for prices and FX rates and 92 days for index readings unless the owner sets another, from 1 to 3660 days
    role: exception
    status: modeled
    span:
      start: 0
      end: 143
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:41.590Z'
id: REQ-valuation-gap-threshold-default
type: req
---
A kind's gap threshold must be 7 days for prices and FX rates and 92 days for index readings unless the owner sets another, from 1 to 3660 days.

## Context

D-94 lets the owner set the gap threshold but states no default and no unit. Claude Code chose one threshold per kind of market data, in days, with these defaults; the author has not reviewed them.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
