---
title: The owner's choice of a series' source must name a source with a value of the series, and a change of source must revalue, from their first transaction, the accounts that held the instrument, or every account for an FX rate or an index.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The owner's choice of a series' source must name a source with a value of the series, and a change of source must revalue, from their first transaction, the accounts that held the instrument, or every account for an FX rate or an index.
logic_claims:
  - CLAIM-ACBE2D562F492D22
semantic_clauses:
  - The owner's choice of a series' source must name a source with a value of the series, and a change of source must revalue, from their first transaction, the accounts that held the instrument, or every account for an FX rate or an index.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 68aec4a5083992c1bafd569b929e942422e19a813a089b3dcd417805984efa88
semantic_inventory:
  - claim_key: CLAIM-ACBE2D562F492D22
    claim_text: The owner's choice of a series' source must name a source with a value of the series, and a change of source must revalue, from their first transaction, the accounts that held the instrument, or every account for an FX rate or an index
    role: normative
    status: modeled
    span:
      start: 0
      end: 235
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:31.023Z'
id: REQ-valuation-series-source-owner-choice
type: req
---
The owner's choice of a series' source must name a source with a value of the series, and a change of source must revalue, from their first transaction, the accounts that held the instrument, or every account for an FX rate or an index.

## Context

D-94 lets the owner change the source of one instrument, currency pair or index. Which sources the owner may choose and what a change revalues are not stated; this is the narrow reading Claude Code built, not reviewed by the author.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
